import { useCallback } from "react";

import { useAtomValue, useSetAtom } from "jotai";

import {
  browserNotificationPermissionAtom,
  requestBrowserNotificationPermissionAtom,
  notificationSettingsAtomReadOnly,
} from "@/providers/store/notification";

export const useBrowserNotification = () => {
  const permission = useAtomValue(browserNotificationPermissionAtom);
  const settings = useAtomValue(notificationSettingsAtomReadOnly);
  const requestPermission = useSetAtom(requestBrowserNotificationPermissionAtom);

  const showNotification = useCallback(
    (title: string, options?: NotificationOptions) => {
      // 通知が無効化されている場合は何もしない
      if (!settings.browserNotificationEnabled) {
        return;
      }

      // 許可されていない場合は何もしない
      if (permission !== "granted") {
        return;
      }

      // ブラウザが通知をサポートしていない場合は何もしない
      if (!("Notification" in window)) {
        return;
      }

      try {
        const notification = new Notification(title, {
          icon: "/favicon.ico",
          badge: "/favicon.ico",
          ...options,
        });

        // 通知をクリックしたときにウィンドウをフォーカス
        notification.onclick = () => {
          window.focus();
          notification.close();
        };

        // 5秒後に自動で閉じる
        setTimeout(() => {
          notification.close();
        }, 5000);

        return notification;
      } catch (error) {
        console.error("通知の表示に失敗しました:", error);
      }
    },
    [permission, settings.browserNotificationEnabled]
  );

  const showMentionNotification = useCallback(
    (channelName: string, userName: string, message: string) => {
      return showNotification(`${userName} が #${channelName} でメンションしました`, {
        body: message,
        tag: `mention-${channelName}`,
        requireInteraction: true,
      });
    },
    [showNotification]
  );

  const showMessageNotification = useCallback(
    (channelName: string, userName: string, message: string) => {
      return showNotification(`${userName} が #${channelName} にメッセージを投稿しました`, {
        body: message,
        tag: `message-${channelName}`,
      });
    },
    [showNotification]
  );

  const showReactionNotification = useCallback(
    (channelName: string, userName: string, emoji: string) => {
      return showNotification(`${userName} が #${channelName} でリアクションしました`, {
        body: `${emoji} を追加`,
        tag: `reaction-${channelName}`,
      });
    },
    [showNotification]
  );

  return {
    permission,
    settings,
    requestPermission,
    showNotification,
    showMentionNotification,
    showMessageNotification,
    showReactionNotification,
  };
};
