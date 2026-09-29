import {
  isPermissionGranted,
  requestPermission,
  sendNotification,
} from "@tauri-apps/plugin-notification";

export const isGranted = () => isPermissionGranted();

export const request = async () => (await requestPermission()) === "granted";

// デスクトップの通知はクリックを受け取れないため、押すとアプリが前に出るだけになる
export const show = (title: string, body: string) => {
  sendNotification({ body, title });
};
