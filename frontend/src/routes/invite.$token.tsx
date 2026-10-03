import { createFileRoute } from "@tanstack/react-router";

import { InvitationAccept } from "#/features/auth/components/InvitationAccept";

export const Route = createFileRoute("/invite/$token")({ component: InvitationAccept });
