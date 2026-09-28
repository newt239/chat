import { useEffect } from "react";

import { useAtomValue, useSetAtom } from "jotai";
import { useTranslation } from "react-i18next";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { toDate } from "#/lib/timestamp";
import { userAtom } from "#/providers/store/auth";
import { addNotificationAtom } from "#/providers/store/notification";
import { useWsClient } from "#/providers/ws/useWsClient";

/** WebSocket の新着メッセージから通知を積む。自分の投稿と表示中チャンネルは対象外 */
export const useNotificationSync = (
  workspaceId: string | null,
  currentChannelId: string | null,
) => {
  const { t } = useTranslation();
  const { wsClient } = useWsClient();
  const currentUser = useAtomValue(userAtom);
  const addNotification = useSetAtom(addNotificationAtom);
  const { data: channels } = useChannels(workspaceId);
  const displayName = useDisplayName();

  useEffect(() => {
    if (!wsClient || workspaceId === null) {
      return undefined;
    }

    return wsClient.on("newMessage", ({ channelId, message }) => {
      if (
        message === undefined ||
        message.userId === currentUser?.id ||
        channelId === currentChannelId
      ) {
        return;
      }

      const isMention = message.mentions.some((mention) => mention.userId === currentUser?.id);
      const channelName = channels?.find((channel) => channel.id === channelId)?.name ?? "";
      const userName = displayName(message.userId, message.user?.displayName ?? "");

      addNotification({
        channelId,
        channelName,
        id: message.id,
        isRead: false,
        message: message.body,
        messageId: message.id,
        timestamp: toDate(message.createdAt),
        title: isMention ? t("notification.mentionTitle", { name: userName }) : channelName,
        type: isMention ? "mention" : "message",
        userId: message.userId,
        userName,
        workspaceId,
      });
    });
  }, [
    wsClient,
    workspaceId,
    currentChannelId,
    currentUser?.id,
    channels,
    addNotification,
    t,
    displayName,
  ]);
};
