import { createFileRoute } from "@tanstack/react-router";

import { WorkspaceIndexPage } from "#/features/layout/components/WorkspaceIndexPage";

export const Route = createFileRoute("/app/$workspaceId/")({ component: WorkspaceIndexPage });
