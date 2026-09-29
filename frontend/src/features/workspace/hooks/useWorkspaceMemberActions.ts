import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

/** ワークスペースメンバーの削除・ロール変更をまとめて提供する */
export const useWorkspaceMemberActions = () => {
  const queryClient = useQueryClient();
  const onSuccess = async (_: unknown, { workspaceId }: { workspaceId?: string }) => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        input: { workspaceId },
        schema: WorkspaceService.method.listMembers,
      }),
    });
  };

  const remove = useMutation(WorkspaceService.method.removeMember, { onSuccess });
  const updateRole = useMutation(WorkspaceService.method.updateMemberRole, { onSuccess });

  return { remove, updateRole };
};
