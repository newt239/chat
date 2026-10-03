import { isTauri } from "#/lib/platform/platform";

// Badging API がなければ何もしない
export const setAppBadge = async (count: number) => {
  if (isTauri) {
    const { getCurrentWindow } = await import("@tauri-apps/api/window");
    // Windows では対応していないため失敗しても無視する
    await getCurrentWindow()
      .setBadgeCount(count > 0 ? count : undefined)
      .catch(() => {});
    return;
  }
  if (!("setAppBadge" in navigator)) {
    return;
  }
  await (count > 0 ? navigator.setAppBadge(count) : navigator.clearAppBadge());
};
