import { Outlet } from "@tanstack/react-router";

import { useSyncPreferences } from "#/features/settings/hooks/usePreferences";

export const AppLayout = () => {
  useSyncPreferences();
  return <Outlet />;
};
