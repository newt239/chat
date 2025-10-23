import { createFileRoute } from "@tanstack/react-router";

import { WorkspaceSelection } from "@/features/workspace/components/WorkspaceSelection";

export const Route = createFileRoute("/app/")({
  component: WorkspaceSelection,
});
