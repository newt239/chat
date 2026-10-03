import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { AdminService } from "#/gen/chat/v1/admin_service_pb";
import { InvitationService } from "#/gen/chat/v1/invitation_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

// 既存ユーザーは招待と同時にメンバーへ追加されるため、メンバー一覧も取り直す
const affectedServices = [InvitationService, WorkspaceService, AdminService];

export const useInvitationActions = () => {
  const queryClient = useQueryClient();
  const onSuccess = () =>
    Promise.all(
      affectedServices.map((schema) =>
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({ cardinality: undefined, schema }),
        }),
      ),
    );

  return {
    create: useMutation(InvitationService.method.createInvitation, { onSuccess }),
    revoke: useMutation(InvitationService.method.revokeInvitation, { onSuccess }),
  };
};
