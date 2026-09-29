import { createFileRoute } from "@tanstack/react-router";

import { InsightsPage } from "#/features/insights/components/InsightsPage";

export const Route = createFileRoute("/app/$workspaceId/insights")({ component: InsightsPage });
