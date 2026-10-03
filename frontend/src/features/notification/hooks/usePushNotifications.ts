import { useEffect, useEffectEvent } from "react";

import { useMutation } from "@connectrpc/connect-query";
import { useAtom } from "jotai";

import { notificationPreferencesAtom } from "#/features/notification/atoms";
import {
  isNotificationGranted,
  requestNotificationPermission,
} from "#/features/notification/utils/notify";
import {
  isPushSupported,
  registerPush,
  unregisterPush,
} from "#/features/notification/utils/pushMessaging";
import { NotificationService, PushPlatform } from "#/gen/chat/v1/notification_service_pb";

/** この端末へのプッシュ通知の登録と解除。トークンはサーバーと端末の両方に持つ */
export const usePushNotifications = () => {
  const [preferences, setPreferences] = useAtom(notificationPreferencesAtom);
  const { mutateAsync: register } = useMutation(NotificationService.method.registerPushToken);
  const { mutateAsync: unregister } = useMutation(NotificationService.method.unregisterPushToken);

  const setToken = (pushToken: string | null) => {
    setPreferences((prev) => ({ ...prev, pushToken }));
  };

  const registerToken = async () => {
    const token = await registerPush();
    await register({ platform: PushPlatform.WEB, token });
    return token;
  };

  // 許可を求められるのはユーザー操作の中だけなので、スイッチを押したときに呼ぶ
  const enable = async () => {
    if (!(await requestNotificationPermission())) {
      return false;
    }
    setToken(await registerToken());
    return true;
  };

  const disable = async () => {
    const token = preferences.pushToken;
    setToken(null);
    if (token !== null) {
      await Promise.all([unregister({ token }), unregisterPush()]);
    }
  };

  // 起動のたびに登録し直す。トークンが変わっていたら古いものを外す
  const refresh = async () => {
    const previous = preferences.pushToken;
    if (previous === null) {
      return;
    }
    if (!(await isNotificationGranted())) {
      await disable();
      return;
    }
    const token = await registerToken();
    if (token !== previous) {
      setToken(token);
      await unregister({ token: previous });
    }
  };

  return {
    disable,
    enable,
    enabled: preferences.pushToken !== null,
    refresh,
    supported: isPushSupported(),
  };
};

/** ワークスペースを開いたときに一度だけ登録を更新する */
export const useSyncPushToken = () => {
  const { refresh, supported } = usePushNotifications();
  const refreshOnStart = useEffectEvent(async () => {
    try {
      await refresh();
    } catch (error) {
      console.warn("プッシュ通知の登録を更新できませんでした", error);
    }
  });

  useEffect(() => {
    if (supported) {
      void refreshOnStart();
    }
  }, [supported]);
};
