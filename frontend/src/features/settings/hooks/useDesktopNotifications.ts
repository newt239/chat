import { useEffect, useEffectEvent } from "react";

import { useAtomValue } from "jotai";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { useDMs } from "#/features/dm/hooks/useDM";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMentionDirectory } from "#/features/message/hooks/useMentionDirectory";
import { notificationPreferencesAtom } from "#/features/settings/atoms";
import { isNotificationSupported, showNotification } from "#/features/settings/utils/notify";
import { usePreferences } from "#/hooks/usePreferences";
import { navigateTo } from "#/lib/navigation";
import { myUserIdAtom } from "#/providers/store/auth";
import { useWsClient } from "#/providers/ws/useWsClient";

import type { MessageEvent } from "#/gen/chat/v1/event_pb";

/** 設定に従って新着メッセージを OS の通知で知らせる。ミュート中と表示中のチャンネルは除く。プッシュ通知と同じ tag で出し、二重にならないようにする */
export const useDesktopNotifications = (workspaceId: string, currentChannelId: string | null) => {
  const { toText } = useMentionDirectory();
  const wsClient = useWsClient();
  const { desktop, pushToken } = useAtomValue(notificationPreferencesAtom);
  const level = usePreferences().notificationLevel;
  const myId = useAtomValue(myUserIdAtom);
  const { data: channels } = useChannels(workspaceId);
  const { data: dms } = useDMs(workspaceId);
  const displayName = useDisplayName();

  // 一覧や表示名が変わるたびに購読し直さないよう、通知の判断はイベントの中で最新の値を読む
  const notify = useEffectEvent(({ channelId, message }: MessageEvent) => {
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

  useEffect(() => {
    if (!wsClient || !desktop || level === "none" || !isNotificationSupported()) {
      return undefined;
    }
    return wsClient.on("newMessage", notify);
  }, [wsClient, desktop, level]);
};
