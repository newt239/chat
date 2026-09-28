import { createFileRoute } from "@tanstack/react-router";

import { WorkspaceIndexPage } from "#/pages/WorkspaceIndexPage";

export const Route = createFileRoute("/app/$workspaceId/")({ component: WorkspaceIndexPage });
