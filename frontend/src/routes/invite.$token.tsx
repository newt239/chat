import { createFileRoute } from "@tanstack/react-router";

import { InvitePage } from "#/pages/InvitePage";

export const Route = createFileRoute("/invite/$token")({ component: InvitePage });
