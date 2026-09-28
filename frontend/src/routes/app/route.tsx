import { createFileRoute, redirect } from "@tanstack/react-router";

import { ResponsiveLayout } from "#/features/layout/components/ResponsiveLayout";
import { store } from "#/providers/store";
import { isAuthenticatedAtom } from "#/providers/store/auth";

// 親の beforeLoad が redirect を throw すると子は評価されないため、ガードはここだけで足りる
export const Route = createFileRoute("/app")({
  beforeLoad: () => {
    if (!store.get(isAuthenticatedAtom)) {
      throw redirect({ to: "/login" });
    }
  },
  component: ResponsiveLayout,
});
