import { createFileRoute } from "@tanstack/react-router";

import { UserGroupListPage } from "#/features/userGroup/components/UserGroupListPage";

export const Route = createFileRoute("/app/$workspaceId/groups")({ component: UserGroupListPage });
