import { useQuery } from "@connectrpc/connect-query";

import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

// GetWorkspace の role はリクエストしたユーザー自身のロール
export const useMyWorkspaceRole = (workspaceId: string) =>
  useQuery(
    WorkspaceService.method.getWorkspace,
    { workspaceId },
    { select: (res) => res.workspace?.role },
  );
