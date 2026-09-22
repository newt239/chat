import { useEffect } from "react";

import { useAtomValue, useSetAtom } from "jotai";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { userAtom } from "#/providers/store/auth";
import { addNotificationAtom } from "#/providers/store/notification";
import { useWsClient } from "#/providers/ws/useWsClient";

/** WebSocket の新着メッセージから通知を積む。自分の投稿と表示中チャンネルは対象外 */
export const useNotificationSync = (
  workspaceId: string | null,
  currentChannelId: string | null,
) => {
  const { wsClient } = useWsClient();
  const currentUser = useAtomValue(userAtom);
  const addNotification = useSetAtom(addNotificationAtom);
  const { data: channels } = useChannels(workspaceId);

  useEffect(() => {
    if (!wsClient || workspaceId === null) {
      return undefined;
    }

    return wsClient.on("new_message", ({ channel_id, message }) => {
      if (message.user.id === currentUser?.id || channel_id === currentChannelId) {
        return;
      }

      const isMention =
        message.mentions?.some((mention) => mention.userId === currentUser?.id) === true;
      const channelName = channels?.find((channel) => channel.id === channel_id)?.name ?? "";

      addNotification({
        channelId: channel_id,
        channelName,
        id: message.id,
        isRead: false,
        message: message.body,
        messageId: message.id,
        timestamp: new Date(message.createdAt),
        title: isMention ? `${message.user.displayName} さんからのメンション` : channelName,
        type: isMention ? "mention" : "message",
        userId: message.user.id,
        userName: message.user.displayName,
        workspaceId,
      });
    });
  }, [wsClient, workspaceId, currentChannelId, currentUser?.id, channels, addNotification]);
};
