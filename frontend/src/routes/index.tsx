import { createFileRoute, redirect } from "@tanstack/react-router";

import { useAuthStore } from "@/lib/store/auth";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    const isAuthenticated = useAuthStore.getState().isAuthenticated;
    if (isAuthenticated) {
      throw redirect({ to: "/app" });
    }
    throw redirect({ to: "/login" });
  },
});
