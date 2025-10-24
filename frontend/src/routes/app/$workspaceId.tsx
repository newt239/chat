import { createFileRoute, redirect } from "@tanstack/react-router";

import { WorkspaceComponent } from "@/features/workspace/components/WorkspaceComponent";
import { store } from "@/lib/store";
import { isAuthenticatedAtom } from "@/lib/store/auth";

const WorkspaceRouteComponent = () => {
  const { workspaceId } = Route.useParams();
  return <WorkspaceComponent workspaceId={workspaceId} />;
};

export const Route = createFileRoute("/app/$workspaceId")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: WorkspaceRouteComponent,
});
