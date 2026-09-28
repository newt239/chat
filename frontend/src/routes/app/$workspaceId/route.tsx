import { createFileRoute } from "@tanstack/react-router";

import { WorkspaceLayout } from "#/pages/WorkspaceLayout";

export const Route = createFileRoute("/app/$workspaceId")({ component: WorkspaceLayout });
