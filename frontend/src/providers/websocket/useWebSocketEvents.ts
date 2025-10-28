import { useEffect, useCallback } from "react";

import { notifications } from "@mantine/notifications";
import { useQueryClient } from "@tanstack/react-query";
import { useAtomValue, useSetAtom } from "jotai";

import { useWebSocket } from "./WebSocketProvider";

import type { MessageWithUser } from "@/features/message/schemas";

import { useBrowserNotification } from "@/features/notification/hooks/useBrowserNotification";
import {
  addNotificationAtom,
  notificationSettingsAtomReadOnly,
} from "@/providers/store/notification";

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
  has_mention: boolean;
};

export const useWebSocketEvents = () => {
  const { client } = useWebSocket();
  const queryClient = useQueryClient();
  const addNotification = useSetAtom(addNotificationAtom);
  const notificationSettings = useAtomValue(notificationSettingsAtomReadOnly);
  const { showMentionNotification, showMessageNotification } = useBrowserNotification();

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

          // 重複を除去してから新しいメッセージを追加
          const uniqueMessages = existingMessages.filter(
            (msg: MessageWithUser) => msg.id !== message.id
          );

          return {
            ...oldData,
            messages: [...uniqueMessages, message],
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

      // 通知処理（自分が送信したメッセージは除外）
      const currentUserId = localStorage.getItem("currentUserId");
      if (currentUserId && message.user?.id !== currentUserId) {
        // チャンネル名を取得
        const workspaceId = channelId.split("-")[0] || "";
        const channelsData = queryClient.getQueryData(["workspaces", workspaceId, "channels"]);

        if (channelsData && typeof channelsData === "object") {
          const data = channelsData as { channels?: Array<{ id: string; name?: string }> };
          const channels = data?.channels || [];
          const channel = channels.find((ch) => ch.id === channelId);
          const channelName = String(channel?.name || "Unknown");

          // メンション検知（簡易版：メッセージ本文に@が含まれているかチェック）
          const isMention = message.body?.includes("@") || false;

          if (isMention) {
            // メンション通知
            addNotification({
              type: "mention",
              title: `${message.user?.displayName || "Unknown"} が #${channelName} でメンションしました`,
              message: message.body || "",
              channelId,
              channelName,
              messageId: message.id,
              userId: String(message.user?.id || ""),
              userName: String(message.user?.displayName || ""),
              workspaceId,
            });

            // ブラウザ通知
            if (message.user?.displayName && message.body) {
              showMentionNotification(channelName, message.user.displayName, message.body);
            }
          } else if (!notificationSettings.mentionOnly) {
            // 通常のメッセージ通知（設定で有効な場合のみ）
            addNotification({
              type: "message",
              title: `${message.user?.displayName || "Unknown"} が #${channelName} にメッセージを投稿しました`,
              message: message.body || "",
              channelId,
              channelName,
              messageId: message.id,
              userId: String(message.user?.id || ""),
              userName: String(message.user?.displayName || ""),
              workspaceId,
            });

            // ブラウザ通知
            if (message.user?.displayName && message.body) {
              showMessageNotification(channelName, message.user.displayName, message.body);
            }
          }
        }
      }
    },
    [
      queryClient,
      addNotification,
      notificationSettings.mentionOnly,
      showMentionNotification,
      showMessageNotification,
    ]
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

          // 重複を除去してから更新
          const uniqueMessages = oldData.messages.filter(
            (msg: MessageWithUser) => msg.id !== updatedMessage.id
          );

          return {
            ...oldData,
            messages: [...uniqueMessages, updatedMessage],
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
      const hasMention = data.has_mention;

      // チャンネルリストの未読数を更新
      queryClient.setQueryData(
        ["workspaces", channelId.split("-")[0], "channels"],
        (
          oldData:
            | {
                channels: Array<{
                  id: string;
                  unread_count: number;
                  has_mention: boolean;
                  [key: string]: unknown;
                }>;
              }
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
                  has_mention: hasMention,
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
