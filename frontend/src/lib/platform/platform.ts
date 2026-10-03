// ビルド時に決まるので、Web 版のバンドルからは Tauri 向けの分岐が消える
export const isTauri = import.meta.env.MODE === "tauri";

// Tauri の CLI がビルド時に渡す
export const isMobileApp =
  import.meta.env.TAURI_ENV_PLATFORM === "android" || import.meta.env.TAURI_ENV_PLATFORM === "ios";
