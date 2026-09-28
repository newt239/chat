import { atomWithStorage } from "jotai/utils";

export const notificationLevels = ["all", "mentions", "none"] as const;
type NotificationLevel = (typeof notificationLevels)[number];

type NotificationPreferences = {
  // ブラウザのデスクトップ通知を出すか
  desktop: boolean;
  level: NotificationLevel;
};

// 通知の受け取り方は端末ごとに決める（通知の許可がブラウザごとに違うため）
export const notificationPreferencesAtom = atomWithStorage<NotificationPreferences>(
  "notification-preferences",
  { desktop: false, level: "mentions" },
  undefined,
  { getOnInit: true },
);
