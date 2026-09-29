import { createFileRoute } from "@tanstack/react-router";

import { MePage } from "#/features/layout/components/MePage";

export const Route = createFileRoute("/app/$workspaceId/me")({ component: MePage });
