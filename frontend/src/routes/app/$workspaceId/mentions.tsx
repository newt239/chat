import { createFileRoute } from "@tanstack/react-router";

import { MentionsPage } from "#/pages/MentionsPage";

export const Route = createFileRoute("/app/$workspaceId/mentions")({ component: MentionsPage });
