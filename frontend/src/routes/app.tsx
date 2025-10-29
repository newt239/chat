import { createFileRoute, redirect } from "@tanstack/react-router";

import { ResponsiveLayout } from "@/features/layout/components/ResponsiveLayout";
import { store } from "@/providers/store";
import { isAuthenticatedAtom } from "@/providers/store/auth";

export const Route = createFileRoute("/app")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: ResponsiveLayout,
});
