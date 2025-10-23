import { createFileRoute } from "@tanstack/react-router";

import { WorkspaceSelection } from "@/components/WorkspaceSelection";

export const Route = createFileRoute("/app/")({
  component: WorkspaceSelection,
});
