import { skipToken, useQuery } from "@connectrpc/connect-query";

import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

export const useMembers = (workspaceId: string | null) =>
  useQuery(
    WorkspaceService.method.listMembers,
    workspaceId === null ? skipToken : { workspaceId },
    { select: (res) => res.members },
  );
