import { createFileRoute } from "@tanstack/react-router";

import { InvitePage } from "#/features/auth/components/InvitePage";

export const Route = createFileRoute("/invite/$token")({ component: InvitePage });
