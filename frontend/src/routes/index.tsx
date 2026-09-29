import { createFileRoute, redirect } from "@tanstack/react-router";

import { isAuthenticatedAtom } from "#/providers/store/auth";
import { store } from "#/providers/store/store";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    throw redirect({ to: store.get(isAuthenticatedAtom) ? "/app" : "/login" });
  },
});
