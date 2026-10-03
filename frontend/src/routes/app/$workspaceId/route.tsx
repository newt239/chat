import { createFileRoute } from "@tanstack/react-router";

import { WorkspaceLayout } from "#/features/layout/components/WorkspaceLayout";
import { workspaceSearchSchema } from "#/lib/overlaySearch";
import { store } from "#/providers/store/store";
import { lastWorkspaceIdAtom } from "#/providers/store/workspace";

export const Route = createFileRoute("/app/$workspaceId")({
  component: WorkspaceLayout,
  onEnter: ({ params }) => {
    store.set(lastWorkspaceIdAtom, params.workspaceId);
  },
  validateSearch: workspaceSearchSchema,
});
