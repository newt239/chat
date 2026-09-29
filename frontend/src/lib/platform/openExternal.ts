import { isTauri } from "#/lib/platform/platform";

export const openExternal = async (url: string) => {
  if (isTauri) {
    const { openUrl } = await import("@tauri-apps/plugin-opener");
    await openUrl(url);
    return;
  }
  globalThis.open(url, "_blank", "noopener,noreferrer");
};
