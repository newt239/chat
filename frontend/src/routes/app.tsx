import { createFileRoute, redirect, Outlet } from "@tanstack/react-router";

import { AppLayout } from "@/components/layout/AppLayout";
import { useAuthStore } from "@/lib/store/auth";

const AppComponent = () => {
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  );
};

export const Route = createFileRoute("/app")({
  beforeLoad: () => {
    const isAuthenticated = useAuthStore.getState().isAuthenticated;
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: AppComponent,
});
