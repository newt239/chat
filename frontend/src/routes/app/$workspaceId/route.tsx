import { createFileRoute } from "@tanstack/react-router";

import { WorkspaceLayout } from "#/features/layout/components/WorkspaceLayout";
import { workspaceSearchSchema } from "#/features/layout/schemas";

export const Route = createFileRoute("/app/$workspaceId")({
  component: WorkspaceLayout,
  validateSearch: workspaceSearchSchema,
});
