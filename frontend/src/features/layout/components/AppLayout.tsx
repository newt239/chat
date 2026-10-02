import { Outlet } from "@tanstack/react-router";

import { useTimezoneSync } from "#/features/settings/hooks/useTimezoneSync";

export const AppLayout = () => {
  useTimezoneSync();
  return <Outlet />;
};
