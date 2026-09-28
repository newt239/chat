import { createFileRoute, redirect } from "@tanstack/react-router";

import { store } from "#/providers/store";
import { isAuthenticatedAtom } from "#/providers/store/auth";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    throw redirect({ to: store.get(isAuthenticatedAtom) ? "/app" : "/login" });
  },
});
