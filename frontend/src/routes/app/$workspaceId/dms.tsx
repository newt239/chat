import { createFileRoute } from "@tanstack/react-router";

import { DMsPage } from "#/features/channel/components/DMsPage";

export const Route = createFileRoute("/app/$workspaceId/dms")({ component: DMsPage });
