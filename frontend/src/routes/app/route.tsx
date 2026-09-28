import { createFileRoute, redirect } from "@tanstack/react-router";

import { AppLayout } from "#/pages/AppLayout";
import { store } from "#/providers/store";
import { isAuthenticatedAtom } from "#/providers/store/auth";

// 親の beforeLoad が redirect を throw すると子は評価されないため、ガードはここだけで足りる
export const Route = createFileRoute("/app")({
  beforeLoad: () => {
    if (!store.get(isAuthenticatedAtom)) {
      throw redirect({ to: "/login" });
    }
  },
  component: AppLayout,
});
