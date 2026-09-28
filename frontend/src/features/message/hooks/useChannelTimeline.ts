import { useEffect, useMemo, useState } from "react";

import type { MessageWithUser, TimelineItem } from "#/features/message/types";
import type { WsClient } from "#/lib/ws";

type UseChannelTimelineArgs = {
  currentChannelId: string | null;
  wsClient: WsClient | null;
  initialMessages: TimelineItem[] | undefined;
};

const replaceMessage = (items: TimelineItem[], message: MessageWithUser) =>
  items.map((item) =>
    item.type === "user" && item.userMessage?.id === message.id
      ? { ...item, userMessage: message }
      : item,
  );

/** リアクションの追加・削除をタイムライン上のメッセージへ反映する */
const applyReaction = (
  items: TimelineItem[],
  messageId: string,
  update: (reactions: NonNullable<MessageWithUser["reactions"]>) => MessageWithUser["reactions"],
) =>
  items.map((item) => {
    if (item.type !== "user" || item.userMessage?.id !== messageId) {
      return item;
    }
    return {
      ...item,
      userMessage: { ...item.userMessage, reactions: update(item.userMessage.reactions ?? []) },
    };
  });

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
      wsClient.on("new_message", ({ channel_id, message }) => {
        if (!isCurrentChannel(channel_id)) {
          return;
        }
        setTimeline((prev) =>
          prev.some((item) => item.type === "user" && item.userMessage?.id === message.id)
            ? prev
            : [...prev, { createdAt: message.createdAt, type: "user", userMessage: message }],
        );
      }),

      wsClient.on("message_updated", ({ channel_id, message }) => {
        if (isCurrentChannel(channel_id)) {
          setTimeline((prev) => replaceMessage(prev, message));
        }
      }),

      wsClient.on("message_deleted", ({ channel_id, deleteData }) => {
        if (!isCurrentChannel(channel_id)) {
          return;
        }
        setTimeline((prev) =>
          prev.map((item) =>
            item.type === "user" && item.userMessage?.id === deleteData.id
              ? {
                  ...item,
                  userMessage: {
                    ...item.userMessage,
                    deletedAt: deleteData.deleted_at,
                    isDeleted: true,
                  },
                }
              : item,
          ),
        );
      }),

      wsClient.on("system_message_created", ({ channel_id, message }) => {
        if (!isCurrentChannel(channel_id)) {
          return;
        }
        setTimeline((prev) =>
          prev.some((item) => item.type === "system" && item.systemMessage?.id === message.id)
            ? prev
            : [...prev, { createdAt: message.createdAt, systemMessage: message, type: "system" }],
        );
      }),

      wsClient.on("reaction_added", ({ channel_id, emoji, message_id, user_id }) => {
        if (!isCurrentChannel(channel_id)) {
          return;
        }
        setTimeline((prev) =>
          applyReaction(prev, message_id, (reactions) =>
            reactions.some((r) => r.emoji === emoji && r.user.id === user_id)
              ? reactions
              : [
                  ...reactions,
                  {
                    createdAt: new Date().toISOString(),
                    emoji,
                    user: { displayName: "", id: user_id },
                  },
                ],
          ),
        );
      }),

      wsClient.on("reaction_removed", ({ channel_id, emoji, message_id, user_id }) => {
        if (!isCurrentChannel(channel_id)) {
          return;
        }
        setTimeline((prev) =>
          applyReaction(prev, message_id, (reactions) =>
            reactions.filter((r) => !(r.emoji === emoji && r.user.id === user_id)),
          ),
        );
      }),

      wsClient.on("typing", ({ channel_id, user_id }) => {
        if (isCurrentChannel(channel_id)) {
          setTypingUserIds((prev) => (prev.includes(user_id) ? prev : [...prev, user_id]));
        }
      }),

      wsClient.on("stop_typing", ({ channel_id, user_id }) => {
        if (isCurrentChannel(channel_id)) {
          setTypingUserIds((prev) => prev.filter((id) => id !== user_id));
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
      return [] as TimelineItem[];
    }

    const seen = new Set<string>();
    const unique = timeline.filter((item) => {
      const id = item.userMessage?.id ?? item.systemMessage?.id;
      if (id === undefined || seen.has(id)) {
        return id === undefined;
      }
      seen.add(id);
      return true;
    });

    return unique.toSorted(
      (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
    );
  }, [timeline, currentChannelId]);

  return { orderedItems, typingUserIds };
};
