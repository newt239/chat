import { createFileRoute } from "@tanstack/react-router";

import { DraftsPage } from "#/features/draft/components/DraftsPage";

export const Route = createFileRoute("/app/$workspaceId/drafts")({ component: DraftsPage });
