import { getApps, initializeApp } from "firebase/app";
import { getMessaging, onRegistered, register, unregister } from "firebase/messaging";

import { firebaseConfig } from "#/lib/firebaseConfig";

/** Firebase の設定があり、ブラウザが Service Worker とプッシュに対応しているか */
export const isPushSupported = () =>
  firebaseConfig !== null &&
  "serviceWorker" in navigator &&
  "PushManager" in globalThis &&
  "Notification" in globalThis;

const messaging = () => {
  if (firebaseConfig === null) {
    throw new Error("Firebase が設定されていません");
  }
  return getMessaging(getApps()[0] ?? initializeApp(firebaseConfig.options));
};

// この端末を FCM に登録し、送信先の ID（Firebase Installation ID）を返す。ID は register の完了までに onRegistered で届く
export const registerPush = async () => {
  const instance = messaging();
  const registered = new Promise<string>((resolve) => {
    onRegistered(instance, resolve);
  });
  await register(instance, {
    serviceWorkerRegistration: await navigator.serviceWorker.ready,
    vapidKey: firebaseConfig?.vapidKey,
  });
  return registered;
};

export const unregisterPush = () => unregister(messaging());
