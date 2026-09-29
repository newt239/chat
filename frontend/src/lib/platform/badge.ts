// Badging API がなければ何もしない
export const setAppBadge = async (count: number) => {
  if (!("setAppBadge" in navigator)) {
    return;
  }
  await (count > 0 ? navigator.setAppBadge(count) : navigator.clearAppBadge());
};
