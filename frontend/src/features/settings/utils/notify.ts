import { isTauri } from "#/lib/platform/platform";

type ShowNotificationOptions = {
  title: string;
  body: string;
  tag: string;
  onClick: () => void;
};

const loadTauri = () => import("@tauri-apps/plugin-notification");

export const isNotificationSupported = () => isTauri || "Notification" in globalThis;

export const isNotificationGranted = async () => {
  if (isTauri) {
    const { isPermissionGranted } = await loadTauri();
    return isPermissionGranted();
  }
  return isNotificationSupported() && Notification.permission === "granted";
};

// 許可を求められるのはユーザー操作の中だけなので、操作のハンドラから呼ぶ
export const requestNotificationPermission = async () => {
  if (isTauri) {
    const { requestPermission } = await loadTauri();
    return (await requestPermission()) === "granted";
  }
  return (await Notification.requestPermission()) === "granted";
};

// 許可がなければ何もしない
export const showNotification = async ({ title, body, tag, onClick }: ShowNotificationOptions) => {
  if (!(await isNotificationGranted())) {
    return;
  }
  if (isTauri) {
    const { sendNotification } = await loadTauri();
    // デスクトップの通知はクリックを受け取れないため、押すとアプリが前に出るだけになる
    sendNotification({ body, title });
    return;
  }
  const notification = new Notification(title, { body, tag });
  notification.addEventListener("click", onClick);
};
