import { useSyncExternalStore } from "react";

// beforeinstallprompt は画面を開く前に届くため、起動時から受け取って取っておく
let deferred: BeforeInstallPromptEvent | null = null;
const listeners = new Set<() => void>();

const notify = () => {
  for (const listener of listeners) {
    listener();
  }
};

export const listenInstallPrompt = () => {
  globalThis.addEventListener("beforeinstallprompt", (event) => {
    event.preventDefault();
    deferred = event;
    notify();
  });
  globalThis.addEventListener("appinstalled", () => {
    deferred = null;
    notify();
  });
};

const subscribe = (listener: () => void) => {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
};

const isStandalone = () => globalThis.matchMedia("(display-mode: standalone)").matches;
// iOS の Safari は beforeinstallprompt を送らないため、共有メニューからの追加を案内する
const isIOS = () => /iPhone|iPad|iPod/.test(navigator.userAgent);

/** インストールの方法。すでにインストール済みか方法がなければ null */
export const useInstallPrompt = () =>
  useSyncExternalStore(subscribe, () => {
    if (deferred) {
      return deferred;
    }
    return isIOS() && !isStandalone() ? "ios" : null;
  });

export const promptInstall = async (event: BeforeInstallPromptEvent) => {
  await event.prompt();
  deferred = null;
  notify();
};
