import { useQuery } from "@connectrpc/connect-query";

import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { isAdminRole } from "#/lib/isAdminRole";

// GetWorkspace の role はリクエストしたユーザー自身のロール
export const useIsWorkspaceAdmin = (workspaceId: string) =>
  useQuery(
    WorkspaceService.method.getWorkspace,
    { workspaceId },
    { select: (res) => isAdminRole(res.workspace?.role) },
  ).data ?? false;
