import { atomWithStorage } from "jotai/utils";

type NotificationPreferences = {
  // ブラウザのデスクトップ通知を出すか。通知の許可がブラウザごとに違うため端末ごとに決める
  desktop: boolean;
  // この端末で登録したプッシュ通知のトークン。登録していなければ null
  pushToken: string | null;
};

export const notificationPreferencesAtom = atomWithStorage<NotificationPreferences>(
  "notification-preferences",
  { desktop: false, pushToken: null },
  undefined,
  { getOnInit: true },
);
