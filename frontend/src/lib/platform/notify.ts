import { isTauri } from "#/lib/platform/platform";

type ShowNotificationOptions = {
  title: string;
  body: string;
  tag: string;
  onClick: () => void;
};

const loadTauri = () => import("#/lib/platform/tauri/notification");

export const isNotificationSupported = () => isTauri || "Notification" in globalThis;

export const isNotificationGranted = async () => {
  if (isTauri) {
    const { isGranted } = await loadTauri();
    return isGranted();
  }
  return isNotificationSupported() && Notification.permission === "granted";
};

// 許可を求められるのはユーザー操作の中だけなので、操作のハンドラから呼ぶ
export const requestNotificationPermission = async () => {
  if (isTauri) {
    const { request } = await loadTauri();
    return request();
  }
  return (await Notification.requestPermission()) === "granted";
};

// 許可がなければ何もしない
export const showNotification = async ({ title, body, tag, onClick }: ShowNotificationOptions) => {
  if (!(await isNotificationGranted())) {
    return;
  }
  if (isTauri) {
    const { show } = await loadTauri();
    show(title, body);
    return;
  }
  const notification = new Notification(title, { body, tag });
  notification.addEventListener("click", onClick);
};
