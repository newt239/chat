import { createFileRoute } from "@tanstack/react-router";

import { DMsPage } from "#/features/dm/components/DMsPage";

export const Route = createFileRoute("/app/$workspaceId/dms")({ component: DMsPage });
