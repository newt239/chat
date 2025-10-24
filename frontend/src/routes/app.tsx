import { createFileRoute, redirect, Outlet } from "@tanstack/react-router";

import { AppLayout } from "@/components/layout/AppLayout";
import { store } from "@/lib/store";
import { isAuthenticatedAtom } from "@/lib/store/auth";

const AppComponent = () => {
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  );
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
