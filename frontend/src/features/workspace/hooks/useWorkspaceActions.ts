import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { workspaceServiceKey } from "#/features/workspace/hooks/useWorkspace";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

/** ワークスペース自体の更新・削除を提供する */
export const useWorkspaceActions = () => {
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await queryClient.invalidateQueries({ queryKey: workspaceServiceKey });
  };

  const update = useMutation(WorkspaceService.method.updateWorkspace, { onSuccess });
  const remove = useMutation(WorkspaceService.method.deleteWorkspace, { onSuccess });

  return { remove, update };
};
