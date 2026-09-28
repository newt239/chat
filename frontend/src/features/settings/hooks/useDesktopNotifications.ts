import { useEffect } from "react";

import { useAtomValue } from "jotai";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { useDMs } from "#/features/dm/hooks/useDM";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { navigateTo } from "#/lib/navigation";
import { userAtom } from "#/providers/store/auth";
import { notificationPreferencesAtom } from "#/providers/store/notificationPreferences";
import { useWsClient } from "#/providers/ws/useWsClient";

/** 設定に従って新着メッセージをブラウザの通知で知らせる。ミュート中と表示中のチャンネルは除く */
export const useDesktopNotifications = (workspaceId: string, currentChannelId: string | null) => {
  const { wsClient } = useWsClient();
  const { desktop, level } = useAtomValue(notificationPreferencesAtom);
  const myId = useAtomValue(userAtom)?.id;
  const { data: channels } = useChannels(workspaceId);
  const { data: dms } = useDMs(workspaceId);
  const displayName = useDisplayName();

  useEffect(() => {
    if (!wsClient || !desktop || level === "none" || !("Notification" in globalThis)) {
      return undefined;
    }
    return wsClient.on("newMessage", ({ channelId, message }) => {
      if (
        message === undefined ||
        message.userId === myId ||
        (channelId === currentChannelId && document.visibilityState === "visible") ||
        Notification.permission !== "granted"
      ) {
        return;
      }
      const channel = channels?.find((candidate) => candidate.id === channelId);
      const dm = dms?.find((candidate) => candidate.id === channelId);
      if ((!channel && !dm) || channel?.isMuted || dm?.isMuted) {
        return;
      }
      const isMention = message.mentions.some((mention) => mention.userId === myId);
      if (level === "mentions" && !isMention && !dm) {
        return;
      }
      const author = displayName(message.userId, message.user?.displayName ?? "");
      const notification = new Notification(channel ? `${author} · #${channel.name}` : author, {
        body: message.body,
        tag: message.id,
      });
      notification.addEventListener("click", () => {
        navigateTo({
          params: { channelId, workspaceId },
          search: { message: message.id },
          to: "/app/$workspaceId/$channelId",
        });
      });
    });
  }, [wsClient, desktop, level, myId, channels, dms, currentChannelId, workspaceId, displayName]);
};
