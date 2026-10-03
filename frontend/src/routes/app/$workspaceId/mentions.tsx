import { createFileRoute } from "@tanstack/react-router";

import { MentionsPage } from "#/features/inbox/components/MentionsPage";

export const Route = createFileRoute("/app/$workspaceId/mentions")({ component: MentionsPage });
