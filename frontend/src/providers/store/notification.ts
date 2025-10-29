import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

// 通知の種類
type NotificationType = "mention" | "message" | "reaction";

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
};

// 通知履歴の状態管理
const notificationsAtom = atomWithStorage<NotificationItem[]>("notifications", []);

// 通知一覧の読み取り用 Atom（読み取り専用に公開）
export const notificationItemsAtom = atom((get) => get(notificationsAtom));

// 未読通知数の計算
export const unreadNotificationCountAtom = atom((get) => {
  const notifications = get(notificationsAtom);
  return notifications.filter((notification) => !notification.isRead).length;
});

// 通知を既読にするAtom
export const markNotificationAsReadAtom = atom(null, (get, set, notificationId: string) => {
  const notifications = get(notificationsAtom);
  const updated = notifications.map((notification) =>
    notification.id === notificationId ? { ...notification, isRead: true } : notification
  );
  set(notificationsAtom, updated);
});

// 通知を削除するAtom
export const removeNotificationAtom = atom(null, (get, set, notificationId: string) => {
  const notifications = get(notificationsAtom);
  const updated = notifications.filter((notification) => notification.id !== notificationId);
  set(notificationsAtom, updated);
});
