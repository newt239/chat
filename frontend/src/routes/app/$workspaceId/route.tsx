import { createFileRoute } from "@tanstack/react-router";

import { workspaceSearchSchema } from "#/features/layout/schemas";
import { WorkspaceLayout } from "#/pages/WorkspaceLayout";

export const Route = createFileRoute("/app/$workspaceId")({
  component: WorkspaceLayout,
  validateSearch: workspaceSearchSchema,
});
