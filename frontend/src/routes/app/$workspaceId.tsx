import { createFileRoute, redirect, Outlet } from "@tanstack/react-router";

import { store } from "@/providers/store";
import { isAuthenticatedAtom } from "@/providers/store/auth";

export const Route = createFileRoute("/app/$workspaceId")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: () => <Outlet />,
});
