import { createFileRoute } from "@tanstack/react-router";

import { JoinPage } from "#/features/auth/components/JoinPage";

export const Route = createFileRoute("/join/$workspaceId")({ component: JoinPage });
