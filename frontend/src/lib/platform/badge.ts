import { isTauri } from "#/lib/platform/platform";

// Badging API がなければ何もしない
export const setAppBadge = async (count: number) => {
  if (isTauri) {
    const { setBadgeCount } = await import("#/lib/platform/tauri/badge");
    await setBadgeCount(count);
    return;
  }
  if (!("setAppBadge" in navigator)) {
    return;
  }
  await (count > 0 ? navigator.setAppBadge(count) : navigator.clearAppBadge());
};
