import { createFileRoute } from "@tanstack/react-router";

import { ActivityPage } from "#/features/layout/components/ActivityPage";

export const Route = createFileRoute("/app/$workspaceId/activity")({ component: ActivityPage });
