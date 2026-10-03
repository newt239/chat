import { useEffect } from "react";

import { setAppBadge } from "#/features/layout/utils/appBadge";

import { useUnreadSummary } from "./useUnreadSummary";

// インストールしたアプリのアイコンに未読数を出す
export const useAppBadge = (workspaceId: string) => {
  const { activityUnread, dmUnread } = useUnreadSummary(workspaceId);
  const count = activityUnread + dmUnread;

  useEffect(() => {
    void setAppBadge(count);
  }, [count]);
};
