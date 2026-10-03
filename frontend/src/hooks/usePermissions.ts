import { skipToken, useQuery } from "@connectrpc/connect-query";

import { PermissionService } from "#/gen/chat/v1/permission_service_pb";

export const usePermissions = (workspaceId: string | null) =>
  useQuery(
    PermissionService.method.getPermissions,
    workspaceId === null ? skipToken : { workspaceId },
  );
