import { createFileRoute, redirect } from "@tanstack/react-router";

import { WorkspaceComponent } from "@/features/workspace/components/WorkspaceComponent";
import { store } from "@/providers/store";
import { isAuthenticatedAtom } from "@/providers/store/auth";

export const Route = createFileRoute("/app/$workspaceId")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: WorkspaceComponent,
});
