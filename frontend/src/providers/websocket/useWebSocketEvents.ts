import { useEffect, useCallback } from "react";

import { notifications } from "@mantine/notifications";
import { useQueryClient } from "@tanstack/react-query";

import { useWebSocket } from "./WebSocketProvider";

import type { MessageWithUser } from "@/features/message/schemas";

type NewMessagePayload = {
  channel_id: string;
  message: MessageWithUser;
};

type MessageUpdatedPayload = {
  channel_id: string;
  message: MessageWithUser;
};

type MessageDeletedPayload = {
  channel_id: string;
  delete_data: {
    message_id: string;
  };
};

type UnreadCountPayload = {
  channel_id: string;
  unread_count: number;
};

export const useWebSocketEvents = () => {
  const { client } = useWebSocket();
  const queryClient = useQueryClient();

  const handleNewMessage = useCallback(
    (payload: unknown) => {
      const data = payload as NewMessagePayload;
      const channelId = data.channel_id;
      const message = data.message;

      // メッセージリストを更新
      queryClient.setQueryData(
        ["channels", channelId, "messages"],
        (oldData: { messages: MessageWithUser[]; hasMore: boolean } | undefined) => {
          if (!oldData) return oldData;

          // 既存のメッセージに新しいメッセージを追加
          const existingMessages = oldData.messages || [];
          const messageExists = existingMessages.some(
            (msg: MessageWithUser) => msg.id === message.id
          );

          if (messageExists) {
            return oldData;
          }

          return {
            ...oldData,
            messages: [...existingMessages, message],
          };
        }
      );

      // チャンネルリストの最終メッセージも更新
      queryClient.setQueryData(
        ["workspaces", data.channel_id.split("-")[0], "channels"],
        (oldData: { channels: Array<{ id: string; [key: string]: unknown }> } | undefined) => {
          if (!oldData) return oldData;

          return {
            ...oldData,
            channels: oldData.channels.map((channel: { id: string; [key: string]: unknown }) => {
              if (channel.id === channelId) {
                return {
                  ...channel,
                  last_message: message,
                  last_message_at: message.created_at,
                };
              }
              return channel;
            }),
          };
        }
      );
    },
    [queryClient]
  );

  const handleMessageUpdated = useCallback(
    (payload: unknown) => {
      const data = payload as MessageUpdatedPayload;
      const channelId = data.channel_id;
      const updatedMessage = data.message;

      // メッセージリストを更新
      queryClient.setQueryData(
        ["channels", channelId, "messages"],
        (oldData: { messages: MessageWithUser[]; hasMore: boolean } | undefined) => {
          if (!oldData) return oldData;

          return {
            ...oldData,
            messages: oldData.messages.map((msg: MessageWithUser) =>
              msg.id === updatedMessage.id ? updatedMessage : msg
            ),
          };
        }
      );
    },
    [queryClient]
  );

  const handleMessageDeleted = useCallback(
    (payload: unknown) => {
      const data = payload as MessageDeletedPayload;
      const channelId = data.channel_id;
      const messageId = data.delete_data.message_id;

      // メッセージリストから削除
      queryClient.setQueryData(
        ["channels", channelId, "messages"],
        (oldData: { messages: MessageWithUser[]; hasMore: boolean } | undefined) => {
          if (!oldData) return oldData;

          return {
            ...oldData,
            messages: oldData.messages.filter((msg: MessageWithUser) => msg.id !== messageId),
          };
        }
      );
    },
    [queryClient]
  );

  const handleUnreadCount = useCallback(
    (payload: unknown) => {
      const data = payload as UnreadCountPayload;
      const channelId = data.channel_id;
      const unreadCount = data.unread_count;

      // チャンネルリストの未読数を更新
      queryClient.setQueryData(
        ["workspaces", channelId.split("-")[0], "channels"],
        (
          oldData:
            | { channels: Array<{ id: string; unread_count: number; [key: string]: unknown }> }
            | undefined
        ) => {
          if (!oldData) return oldData;

          return {
            ...oldData,
            channels: oldData.channels.map((channel: { id: string; [key: string]: unknown }) => {
              if (channel.id === channelId) {
                return {
                  ...channel,
                  unread_count: unreadCount,
                };
              }
              return channel;
            }),
          };
        }
      );
    },
    [queryClient]
  );

  const handleError = useCallback((payload: unknown) => {
    const error = payload as { code: string; message: string };
    notifications.show({
      title: "WebSocketエラー",
      message: error.message || "接続エラーが発生しました",
      color: "red",
    });
  }, []);

  useEffect(() => {
    if (!client) return;

    // イベントハンドラーを登録
    client.on("new_message", handleNewMessage);
    client.on("message_updated", handleMessageUpdated);
    client.on("message_deleted", handleMessageDeleted);
    client.on("unread_count", handleUnreadCount);
    client.on("error", handleError);

    return () => {
      // クリーンアップ
      client.off("new_message", handleNewMessage);
      client.off("message_updated", handleMessageUpdated);
      client.off("message_deleted", handleMessageDeleted);
      client.off("unread_count", handleUnreadCount);
      client.off("error", handleError);
    };
  }, [
    client,
    handleNewMessage,
    handleMessageUpdated,
    handleMessageDeleted,
    handleUnreadCount,
    handleError,
  ]);
};
