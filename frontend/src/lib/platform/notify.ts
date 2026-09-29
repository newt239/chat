type ShowNotificationOptions = {
  title: string;
  body: string;
  tag: string;
  onClick: () => void;
};

export const isNotificationSupported = () => "Notification" in globalThis;

export const isNotificationGranted = () =>
  Promise.resolve(isNotificationSupported() && Notification.permission === "granted");

// 許可を求められるのはユーザー操作の中だけなので、操作のハンドラから呼ぶ
export const requestNotificationPermission = async () =>
  (await Notification.requestPermission()) === "granted";

// 許可がなければ何もしない
export const showNotification = async ({ title, body, tag, onClick }: ShowNotificationOptions) => {
  if (!(await isNotificationGranted())) {
    return;
  }
  const notification = new Notification(title, { body, tag });
  notification.addEventListener("click", onClick);
};
