import { createFileRoute } from "@tanstack/react-router";

import { MePage } from "#/pages/MePage";

export const Route = createFileRoute("/app/$workspaceId/me")({ component: MePage });
