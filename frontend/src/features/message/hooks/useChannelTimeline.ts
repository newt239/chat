import { useEffect, useMemo, useState } from "react";

import { create } from "@bufbuild/protobuf";
import { timestampNow } from "@bufbuild/protobuf/wkt";

import { ReactionSchema, TimelineItemSchema } from "#/gen/chat/v1/message_pb";
import { toDate } from "#/lib/timestamp";

import type { Message, Reaction, TimelineItem } from "#/gen/chat/v1/message_pb";
import type { WsClient } from "#/lib/ws";

type UseChannelTimelineArgs = {
  currentChannelId: string | null;
  // 集約表示中の子孫チャンネル。購読してタイムラインに新着を積む
  descendantIds: readonly string[];
  wsClient: WsClient | null;
  initialMessages: TimelineItem[] | undefined;
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

const updateReactions = (
  items: TimelineItem[],
  messageId: string,
  update: (reactions: Reaction[]) => Reaction[],
) =>
  updateUserMessages(
    items,
    (message) => message.id === messageId,
    (message) => ({ ...message, reactions: update(message.reactions) }),
  );

export const useChannelTimeline = ({
  currentChannelId,
  descendantIds,
  wsClient,
  initialMessages,
}: UseChannelTimelineArgs) => {
  const [timeline, setTimeline] = useState<TimelineItem[]>([]);
  const [typingUserIds, setTypingUserIds] = useState<string[]>([]);

  // 初期ロード・チャンネル変更時に初期化
  useEffect(() => {
    setTimeline(initialMessages ?? []);
  }, [initialMessages, currentChannelId]);

  useEffect(() => {
    setTypingUserIds([]);
  }, [currentChannelId]);

  const descendantKey = [...new Set(descendantIds)].toSorted().join(",");

  // WS 購読と join/leave 管理
  useEffect(() => {
    if (!wsClient || !currentChannelId) {
      return undefined;
    }
    const channelIds = [currentChannelId, ...descendantKey.split(",").filter(Boolean)];
    for (const channelId of channelIds) {
      wsClient.joinChannel(channelId);
    }

    const isShownChannel = (channelId: string) => channelIds.includes(channelId);
    const isCurrentChannel = (channelId: string) => channelId === currentChannelId;

    const unsubscribes = [
      wsClient.on("newMessage", ({ channelId, message }) => {
        // スレッドの返信はスレッドパネル側で表示するためタイムラインには積まない
        if (!isShownChannel(channelId) || message === undefined || message.parentId !== undefined) {
          return;
        }
        setTimeline((prev) =>
          prev.some(
            (item) => item.content.case === "userMessage" && item.content.value.id === message.id,
          )
            ? prev
            : [
                ...prev,
                create(TimelineItemSchema, {
                  content: { case: "userMessage", value: message },
                  createdAt: message.createdAt,
                }),
              ],
        );
      }),

      wsClient.on("messageUpdated", ({ channelId, message }) => {
        if (!isShownChannel(channelId) || message === undefined) {
          return;
        }
        setTimeline((prev) =>
          updateUserMessages(
            prev,
            (current) => current.id === message.id,
            () => message,
          ),
        );
      }),

      wsClient.on("messageDeleted", ({ channelId, deletedMessageIds, deletedAt }) => {
        if (!isShownChannel(channelId)) {
          return;
        }
        const deletedIds = new Set(deletedMessageIds);
        setTimeline((prev) =>
          updateUserMessages(
            prev,
            (message) => deletedIds.has(message.id),
            (message) => ({ ...message, deletedAt, isDeleted: true }),
          ),
        );
      }),

      wsClient.on("systemMessageCreated", ({ channelId, message }) => {
        if (!isShownChannel(channelId) || message === undefined) {
          return;
        }
        setTimeline((prev) =>
          prev.some(
            (item) => item.content.case === "systemMessage" && item.content.value.id === message.id,
          )
            ? prev
            : [
                ...prev,
                create(TimelineItemSchema, {
                  content: { case: "systemMessage", value: message },
                  createdAt: message.createdAt,
                }),
              ],
        );
      }),

      wsClient.on("reactionAdded", ({ channelId, emoji, messageId, userId }) => {
        if (!isShownChannel(channelId)) {
          return;
        }
        setTimeline((prev) =>
          updateReactions(prev, messageId, (reactions) =>
            reactions.some((r) => r.emoji === emoji && r.user?.id === userId)
              ? reactions
              : [
                  ...reactions,
                  create(ReactionSchema, {
                    createdAt: timestampNow(),
                    emoji,
                    messageId,
                    user: { id: userId },
                  }),
                ],
          ),
        );
      }),

      wsClient.on("reactionRemoved", ({ channelId, emoji, messageId, userId }) => {
        if (!isShownChannel(channelId)) {
          return;
        }
        setTimeline((prev) =>
          updateReactions(prev, messageId, (reactions) =>
            reactions.filter((r) => !(r.emoji === emoji && r.user?.id === userId)),
          ),
        );
      }),

      wsClient.on("typing", ({ channelId, userId }) => {
        if (isCurrentChannel(channelId)) {
          setTypingUserIds((prev) => (prev.includes(userId) ? prev : [...prev, userId]));
        }
      }),

      wsClient.on("stopTyping", ({ channelId, userId }) => {
        if (isCurrentChannel(channelId)) {
          setTypingUserIds((prev) => prev.filter((id) => id !== userId));
        }
      }),
    ];

    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
      for (const channelId of channelIds) {
        wsClient.leaveChannel(channelId);
      }
    };
  }, [wsClient, currentChannelId, descendantKey]);

  const orderedItems = useMemo(() => {
    if (!currentChannelId) {
      return [];
    }

    const seen = new Set<string>();
    const unique = timeline.filter((item) => {
      const id = item.content.value?.id;
      if (id === undefined || seen.has(id)) {
        return id === undefined;
      }
      seen.add(id);
      return true;
    });

    return unique.toSorted((a, b) => toDate(a.createdAt).getTime() - toDate(b.createdAt).getTime());
  }, [timeline, currentChannelId]);

  return { orderedItems, typingUserIds };
};
