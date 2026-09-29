import { createFileRoute, redirect } from "@tanstack/react-router";

import { AppLayout } from "#/features/layout/components/AppLayout";
import { isAuthenticatedAtom } from "#/providers/store/auth";
import { store } from "#/providers/store/store";

// 親の beforeLoad が redirect を throw すると子は評価されないため、ガードはここだけで足りる
export const Route = createFileRoute("/app")({
  beforeLoad: () => {
    if (!store.get(isAuthenticatedAtom)) {
      throw redirect({ to: "/login" });
    }
  },
  component: AppLayout,
});
