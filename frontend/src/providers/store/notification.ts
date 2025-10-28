import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

// 通知の種類
export type NotificationType = "mention" | "message" | "reaction";

// 通知アイテムの型定義
export type NotificationItem = {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  channelId: string;
  channelName: string;
  messageId?: string;
  userId?: string;
  userName?: string;
  timestamp: Date;
  isRead: boolean;
  workspaceId: string;
};

// 通知設定の型定義
export type NotificationSettings = {
  browserNotificationEnabled: boolean;
  soundEnabled: boolean;
  mentionOnly: boolean;
};

// 通知履歴の状態管理
const notificationsAtom = atomWithStorage<NotificationItem[]>("notifications", []);

// 通知設定の状態管理
const notificationSettingsAtom = atomWithStorage<NotificationSettings>("notification-settings", {
  browserNotificationEnabled: false,
  soundEnabled: true,
  mentionOnly: true,
});

// 未読通知数の計算
export const unreadNotificationCountAtom = atom((get) => {
  const notifications = get(notificationsAtom);
  return notifications.filter((notification) => !notification.isRead).length;
});

// 通知を追加するAtom
export const addNotificationAtom = atom(
  null,
  (get, set, notification: Omit<NotificationItem, "id" | "timestamp" | "isRead">) => {
    const notifications = get(notificationsAtom);
    const newNotification: NotificationItem = {
      ...notification,
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      timestamp: new Date(),
      isRead: false,
    };

    // 重複を避けるため、同じメッセージIDの通知は更新
    const existingIndex = notifications.findIndex(
      (n) => n.messageId === notification.messageId && n.type === notification.type
    );

    if (existingIndex >= 0) {
      const updated = [...notifications];
      updated[existingIndex] = newNotification;
      set(notificationsAtom, updated);
    } else {
      set(notificationsAtom, [newNotification, ...notifications]);
    }
  }
);

// 通知を既読にするAtom
export const markNotificationAsReadAtom = atom(null, (get, set, notificationId: string) => {
  const notifications = get(notificationsAtom);
  const updated = notifications.map((notification) =>
    notification.id === notificationId ? { ...notification, isRead: true } : notification
  );
  set(notificationsAtom, updated);
});

// 全通知を既読にするAtom
export const markAllNotificationsAsReadAtom = atom(null, (get, set) => {
  const notifications = get(notificationsAtom);
  const updated = notifications.map((notification) => ({
    ...notification,
    isRead: true,
  }));
  set(notificationsAtom, updated);
});

// 通知を削除するAtom
export const removeNotificationAtom = atom(null, (get, set, notificationId: string) => {
  const notifications = get(notificationsAtom);
  const updated = notifications.filter((notification) => notification.id !== notificationId);
  set(notificationsAtom, updated);
});

// 古い通知を削除するAtom（30日以上前の通知を削除）
export const cleanupOldNotificationsAtom = atom(null, (get, set) => {
  const notifications = get(notificationsAtom);
  const thirtyDaysAgo = new Date();
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

  const updated = notifications.filter((notification) => notification.timestamp > thirtyDaysAgo);
  set(notificationsAtom, updated);
});

// 通知設定を更新するAtom
export const updateNotificationSettingsAtom = atom(
  null,
  (get, set, settings: Partial<NotificationSettings>) => {
    const currentSettings = get(notificationSettingsAtom);
    set(notificationSettingsAtom, { ...currentSettings, ...settings });
  }
);

// 通知を取得するAtom（読み取り専用）
export const notificationsAtomReadOnly = atom((get) => get(notificationsAtom));

// 通知設定を取得するAtom（読み取り専用）
export const notificationSettingsAtomReadOnly = atom((get) => get(notificationSettingsAtom));

// ブラウザ通知の許可状態を管理するAtom
export const browserNotificationPermissionAtom = atom<NotificationPermission>("default");

// ブラウザ通知の許可をリクエストするAtom
export const requestBrowserNotificationPermissionAtom = atom(null, async (_get, set) => {
  if (!("Notification" in window)) {
    console.warn("このブラウザは通知をサポートしていません");
    return false;
  }

  try {
    const permission = await Notification.requestPermission();
    set(browserNotificationPermissionAtom, permission);
    return permission === "granted";
  } catch (error) {
    console.error("通知の許可リクエストに失敗しました:", error);
    return false;
  }
});
