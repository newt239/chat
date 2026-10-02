import { useEffect, useMemo } from "react";

import { create } from "@bufbuild/protobuf";
import { timestampNow } from "@bufbuild/protobuf/wkt";
import { useQueryClient } from "@tanstack/react-query";

import { MessagePinSchema, ReactionSchema, TimelineItemSchema } from "#/gen/chat/v1/message_pb";
import { useWsClient } from "#/providers/ws/useWsClient";

import { messagePagesKey } from "./useMessagePages";

import type { Message, Reaction, TimelineItem } from "#/gen/chat/v1/message_pb";
import type { ListMessagesResponse } from "#/gen/chat/v1/message_service_pb";

import type { InfiniteData } from "@tanstack/react-query";

type UseChannelTimelineArgs = {
  channelId: string;
  includeDescendants: boolean;
  // 集約表示中の子孫チャンネル。購読してタイムラインに新着を積む
  descendantIds: readonly string[];
  // useMessagePages の項目（新しい順）
  items: TimelineItem[] | undefined;
};

/** ユーザーメッセージの項目だけを更新する */
const updateUserMessages = (
  items: TimelineItem[],
  shouldUpdate: (message: Message) => boolean,
  update: (message: Message) => Message,
) =>
  items.map((item) =>
    item.content.case === "userMessage" && shouldUpdate(item.content.value)
      ? { ...item, content: { case: "userMessage" as const, value: update(item.content.value) } }
      : item,
  );

const updateMessage = (
  items: TimelineItem[],
  messageId: string,
  update: (message: Message) => Message,
) => updateUserMessages(items, (message) => message.id === messageId, update);

const updateReactions = (
  items: TimelineItem[],
  messageId: string,
  update: (reactions: Reaction[]) => Reaction[],
) =>
  updateMessage(items, messageId, (message) => ({
    ...message,
    reactions: update(message.reactions),
  }));

/** WebSocket の差分を読み込み済みのページに当て、表示用に古い順へ並べ直す */
export const useChannelTimeline = ({
  channelId,
  includeDescendants,
  descendantIds,
  items,
}: UseChannelTimelineArgs) => {
  const queryClient = useQueryClient();
  const { wsClient } = useWsClient();
  const descendantKey = [...new Set(descendantIds)].toSorted().join(",");

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }
    const channelIds = [channelId, ...descendantKey.split(",").filter(Boolean)];
    for (const id of channelIds) {
      wsClient.joinChannel(id);
    }

    const updatePages = (
      eventChannelId: string,
      update: (items: TimelineItem[], isLatestPage: boolean) => TimelineItem[],
    ) => {
      if (!channelIds.includes(eventChannelId)) {
        return;
      }
      queryClient.setQueriesData<InfiniteData<ListMessagesResponse>>(
        { queryKey: messagePagesKey(channelId, includeDescendants) },
        (data) =>
          data && {
            ...data,
            pages: data.pages.map((page, index) => ({
              ...page,
              // 新しい順に並ぶので先頭のページが最新。最新まで読み込んでいるときだけ新着を足す
              messages: update(page.messages, index === 0 && !page.hasNewer),
            })),
          },
      );
    };

    const prependToLatest = (eventChannelId: string, item: TimelineItem) => {
      const id = item.content.value?.id;
      updatePages(eventChannelId, (current, isLatestPage) =>
        isLatestPage && !current.some((existing) => existing.content.value?.id === id)
          ? [item, ...current]
          : current,
      );
    };

    const unsubscribes = [
      wsClient.on("newMessage", ({ channelId: eventChannelId, message }) => {
        // スレッドの返信はスレッドパネル側で表示するためタイムラインには積まない
        if (message === undefined || message.parentId !== undefined) {
          return;
        }
        prependToLatest(
          eventChannelId,
          create(TimelineItemSchema, {
            content: { case: "userMessage", value: message },
            createdAt: message.createdAt,
          }),
        );
      }),

      wsClient.on("systemMessageCreated", ({ channelId: eventChannelId, message }) => {
        if (message !== undefined) {
          prependToLatest(
            eventChannelId,
            create(TimelineItemSchema, {
              content: { case: "systemMessage", value: message },
              createdAt: message.createdAt,
            }),
          );
        }
      }),

      wsClient.on("messageUpdated", ({ channelId: eventChannelId, message }) => {
        if (message !== undefined) {
          updatePages(eventChannelId, (current) =>
            updateMessage(current, message.id, () => message),
          );
        }
      }),

      wsClient.on(
        "messageDeleted",
        ({ channelId: eventChannelId, deletedMessageIds, deletedAt }) => {
          const deletedIds = new Set(deletedMessageIds);
          updatePages(eventChannelId, (current) =>
            updateUserMessages(
              current,
              (message) => deletedIds.has(message.id),
              (message) => ({ ...message, deletedAt, isDeleted: true }),
            ),
          );
        },
      ),

      wsClient.on(
        "reactionAdded",
        ({ channelId: eventChannelId, emoji, messageId, userId, user, createdAt }) => {
          updatePages(eventChannelId, (current) =>
            updateReactions(current, messageId, (reactions) =>
              reactions.some((r) => r.emoji === emoji && r.user?.id === userId)
                ? reactions
                : [
                    ...reactions,
                    create(ReactionSchema, {
                      createdAt: createdAt ?? timestampNow(),
                      emoji,
                      messageId,
                      user: user ?? { id: userId },
                    }),
                  ],
            ),
          );
        },
      ),

      wsClient.on("reactionRemoved", ({ channelId: eventChannelId, emoji, messageId, userId }) => {
        updatePages(eventChannelId, (current) =>
          updateReactions(current, messageId, (reactions) =>
            reactions.filter((r) => !(r.emoji === emoji && r.user?.id === userId)),
          ),
        );
      }),

      wsClient.on(
        "pinCreated",
        ({ channelId: eventChannelId, messageId, pinnedByUser, pinnedAt }) => {
          updatePages(eventChannelId, (current) =>
            updateMessage(current, messageId, (message) => ({
              ...message,
              pin: create(MessagePinSchema, { pinnedAt, pinnedBy: pinnedByUser }),
            })),
          );
        },
      ),

      wsClient.on("pinDeleted", ({ channelId: eventChannelId, messageId }) => {
        updatePages(eventChannelId, (current) =>
          updateMessage(current, messageId, (message) => ({ ...message, pin: undefined })),
        );
      }),
    ];

    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
      for (const id of channelIds) {
        wsClient.leaveChannel(id);
      }
    };
  }, [wsClient, queryClient, channelId, includeDescendants, descendantKey]);

  // ページの境目で重なった項目を除き、古い順にする
  const orderedItems = useMemo(() => {
    const seen = new Set<string>();
    return (items ?? [])
      .filter((item) => {
        const id = item.content.value?.id;
        if (id === undefined || seen.has(id)) {
          return id === undefined;
        }
        seen.add(id);
        return true;
      })
      .toReversed();
  }, [items]);

  return { orderedItems };
};
