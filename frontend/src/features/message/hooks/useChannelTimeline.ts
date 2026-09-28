import { useEffect, useMemo, useState } from "react";

import { create } from "@bufbuild/protobuf";
import { timestampNow } from "@bufbuild/protobuf/wkt";

import { MessagePinSchema, ReactionSchema, TimelineItemSchema } from "#/gen/chat/v1/message_pb";
import { toDate } from "#/lib/timestamp";

import type { Message, Reaction, TimelineItem } from "#/gen/chat/v1/message_pb";
import type { WsClient } from "#/lib/ws";

type UseChannelTimelineArgs = {
  currentChannelId: string | null;
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

  // WS 購読と join/leave 管理
  useEffect(() => {
    if (!wsClient || !currentChannelId) {
      return undefined;
    }
    wsClient.joinChannel(currentChannelId);

    const isCurrentChannel = (channelId: string) => channelId === currentChannelId;

    const unsubscribes = [
      wsClient.on("newMessage", ({ channelId, message }) => {
        // スレッドの返信はスレッドパネル側で表示するためタイムラインには積まない
        if (
          !isCurrentChannel(channelId) ||
          message === undefined ||
          message.parentId !== undefined
        ) {
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
        if (!isCurrentChannel(channelId) || message === undefined) {
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
        if (!isCurrentChannel(channelId)) {
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
        if (!isCurrentChannel(channelId) || message === undefined) {
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

      wsClient.on("reactionAdded", ({ channelId, emoji, messageId, userId, user, createdAt }) => {
        if (!isCurrentChannel(channelId)) {
          return;
        }
        setTimeline((prev) =>
          updateReactions(prev, messageId, (reactions) =>
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
      }),

      wsClient.on("reactionRemoved", ({ channelId, emoji, messageId, userId }) => {
        if (!isCurrentChannel(channelId)) {
          return;
        }
        setTimeline((prev) =>
          updateReactions(prev, messageId, (reactions) =>
            reactions.filter((r) => !(r.emoji === emoji && r.user?.id === userId)),
          ),
        );
      }),

      wsClient.on("pinCreated", ({ channelId, messageId, pinnedByUser, pinnedAt }) => {
        if (!isCurrentChannel(channelId)) {
          return;
        }
        setTimeline((prev) =>
          updateUserMessages(
            prev,
            (message) => message.id === messageId,
            (message) => ({
              ...message,
              pin: create(MessagePinSchema, { pinnedAt, pinnedBy: pinnedByUser }),
            }),
          ),
        );
      }),

      wsClient.on("pinDeleted", ({ channelId, messageId }) => {
        if (!isCurrentChannel(channelId)) {
          return;
        }
        setTimeline((prev) =>
          updateUserMessages(
            prev,
            (message) => message.id === messageId,
            (message) => ({ ...message, pin: undefined }),
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
      wsClient.leaveChannel(currentChannelId);
    };
  }, [wsClient, currentChannelId]);

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
