import { createFileRoute, redirect } from "@tanstack/react-router";

import { store } from "@/lib/store";
import { isAuthenticatedAtom } from "@/lib/store/auth";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (isAuthenticated) {
      throw redirect({ to: "/app" });
    }
    throw redirect({ to: "/login" });
  },
});
