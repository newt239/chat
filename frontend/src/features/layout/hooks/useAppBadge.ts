import { useEffect } from "react";

import { useUnreadSummary } from "./useUnreadSummary";

// インストールしたアプリのアイコンに未読数を出す。Badging API がなければ何もしない
export const useAppBadge = (workspaceId: string) => {
  const { activityUnread, dmUnread } = useUnreadSummary(workspaceId);
  const count = activityUnread + dmUnread;

  useEffect(() => {
    if (!("setAppBadge" in navigator)) {
      return;
    }
    void (count > 0 ? navigator.setAppBadge(count) : navigator.clearAppBadge());
  }, [count]);
};
