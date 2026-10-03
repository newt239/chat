import { Outlet } from "@tanstack/react-router";

import { useTimezoneSync } from "#/features/layout/hooks/useTimezoneSync";

export const AppLayout = () => {
  useTimezoneSync();
  return <Outlet />;
};
