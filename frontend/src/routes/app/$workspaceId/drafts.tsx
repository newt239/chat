import { createFileRoute } from "@tanstack/react-router";

import { DraftsPage } from "#/pages/DraftsPage";

export const Route = createFileRoute("/app/$workspaceId/drafts")({ component: DraftsPage });
