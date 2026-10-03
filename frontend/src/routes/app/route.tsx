import { createFileRoute, redirect } from "@tanstack/react-router";

import { AppLayout } from "#/features/layout/components/AppLayout";
import { ensureSession } from "#/lib/session";

// 親の beforeLoad が redirect を throw すると子は評価されないため、ガードはここだけで足りる
export const Route = createFileRoute("/app")({
  beforeLoad: async () => {
    if (!(await ensureSession())) {
      throw redirect({ to: "/login" });
    }
  },
  component: AppLayout,
});
