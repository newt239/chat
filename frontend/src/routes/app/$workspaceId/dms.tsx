import { createFileRoute } from "@tanstack/react-router";

import { DMsPage } from "#/pages/DMsPage";

export const Route = createFileRoute("/app/$workspaceId/dms")({ component: DMsPage });
