import { useEffect } from "react";

import { useAtomValue } from "jotai";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { useDMs } from "#/features/dm/hooks/useDM";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMentionDirectory } from "#/features/message/hooks/useMentionDirectory";
import { navigateTo } from "#/lib/navigation";
import { isNotificationSupported, showNotification } from "#/lib/platform/notify";
import { userAtom } from "#/providers/store/auth";
import { notificationPreferencesAtom } from "#/providers/store/notificationPreferences";
import { preferencesAtom } from "#/providers/store/preferences";
import { useWsClient } from "#/providers/ws/useWsClient";

/** 設定に従って新着メッセージを OS の通知で知らせる。ミュート中と表示中のチャンネルは除く。プッシュ通知と同じ tag で出し、二重にならないようにする */
export const useDesktopNotifications = (workspaceId: string, currentChannelId: string | null) => {
  const { toText } = useMentionDirectory();
  const { wsClient } = useWsClient();
  const { desktop, pushToken } = useAtomValue(notificationPreferencesAtom);
  const level = useAtomValue(preferencesAtom).notificationLevel;
  const myId = useAtomValue(userAtom)?.id;
  const { data: channels } = useChannels(workspaceId);
  const { data: dms } = useDMs(workspaceId);
  const displayName = useDisplayName();

  useEffect(() => {
    if (!wsClient || !desktop || level === "none" || !isNotificationSupported()) {
      return undefined;
    }
    return wsClient.on("newMessage", ({ channelId, message }) => {
      if (
        message === undefined ||
        message.userId === myId ||
        (channelId === currentChannelId && document.hasFocus()) ||
        // 裏にいる間はプッシュ通知が届くので、そちらに任せる
        (pushToken !== null && document.visibilityState !== "visible")
      ) {
        return;
      }
      const channel = channels?.find((candidate) => candidate.id === channelId);
      const dm = dms?.find((candidate) => candidate.id === channelId);
      if ((!channel && !dm) || channel?.isMuted || dm?.isMuted) {
        return;
      }
      const isMention =
        message.mentionsChannel ||
        message.mentionsHere ||
        message.mentions.some((mention) => mention.userId === myId);
      if (level === "mentions" && !isMention && !dm) {
        return;
      }
      const author = displayName(message.userId, message.user?.displayName ?? "");
      void showNotification({
        body: toText(message.body),
        onClick: () => {
          navigateTo({
            params: { channelId, workspaceId },
            search: { message: message.id },
            to: "/app/$workspaceId/$channelId",
          });
        },
        tag: message.id,
        title: channel ? `${author} · #${channel.name}` : author,
      });
    });
  }, [
    wsClient,
    desktop,
    pushToken,
    level,
    myId,
    channels,
    dms,
    currentChannelId,
    workspaceId,
    displayName,
    toText,
  ]);
};
