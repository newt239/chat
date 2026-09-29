// ビルド時に決まるので、Web 版のバンドルからは Tauri 向けの分岐が消える
export const isTauri = import.meta.env.MODE === "tauri";

let isMobileApp = false;

/** Tauri のモバイルアプリか。setupPlatform の後でだけ正しい値になる */
export const getIsMobileApp = () => isMobileApp;

/** 描画の前に一度だけ呼ぶ */
export const setupPlatform = async () => {
  if (isTauri) {
    const { setupTauri } = await import("#/lib/platform/tauri/setup");
    ({ isMobileApp } = setupTauri());
    return;
  }
  const [{ listenInstallPrompt }, { registerServiceWorker }] = await Promise.all([
    import("#/features/layout/utils/installPrompt"),
    import("#/lib/registerServiceWorker"),
  ]);
  listenInstallPrompt();
  registerServiceWorker();
};
