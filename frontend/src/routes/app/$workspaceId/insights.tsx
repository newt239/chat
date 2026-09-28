import { createFileRoute } from "@tanstack/react-router";

import { InsightsPage } from "#/pages/InsightsPage";

export const Route = createFileRoute("/app/$workspaceId/insights")({ component: InsightsPage });
