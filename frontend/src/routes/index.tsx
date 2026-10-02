import { createFileRoute, redirect } from "@tanstack/react-router";

import { ensureSession } from "#/lib/session";

export const Route = createFileRoute("/")({
  beforeLoad: async () => {
    throw redirect({ to: (await ensureSession()) ? "/app" : "/login" });
  },
});
