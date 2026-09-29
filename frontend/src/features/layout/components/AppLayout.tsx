import { Outlet } from "@tanstack/react-router";

import { useSyncPreferences } from "#/features/settings/hooks/usePreferences";
import { useTimezoneSync } from "#/features/settings/hooks/useTimezoneSync";

export const AppLayout = () => {
  useSyncPreferences();
  useTimezoneSync();
  return <Outlet />;
};
