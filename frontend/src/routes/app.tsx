import { createFileRoute, redirect } from "@tanstack/react-router";

import { ResponsiveLayout } from "@/features/workspace/components/ResponsiveLayout";
import { store } from "@/providers/store";
import { isAuthenticatedAtom } from "@/providers/store/auth";

const AppComponent = () => {
  return <ResponsiveLayout />;
};

export const Route = createFileRoute("/app")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: AppComponent,
});
