import { createFileRoute, redirect } from "@tanstack/react-router";

import { store } from "@/providers/store";
import { isAuthenticatedAtom } from "@/providers/store/auth";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (isAuthenticated) {
      throw redirect({ to: "/app" });
    }
    throw redirect({ to: "/login" });
  },
});
