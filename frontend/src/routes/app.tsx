import { createFileRoute, redirect, Outlet } from "@tanstack/react-router";

import { Header } from "@/features/workspace/components/Header";
import { store } from "@/lib/store";
import { isAuthenticatedAtom } from "@/lib/store/auth";

const AppComponent = () => {
  return (
    <div className="h-full flex flex-col bg-gray-50">
      <Header />
      <div className="flex-1 min-h-0 p-6">
        <div className="h-full min-h-0">
          <Outlet />
        </div>
      </div>
    </div>
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
