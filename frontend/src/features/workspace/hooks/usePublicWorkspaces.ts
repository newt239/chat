import { useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { workspaceServiceKey } from "#/features/workspace/hooks/useWorkspace";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

/** 参加できる公開ワークスペースの一覧を取得する */
export const usePublicWorkspaces = () =>
  useQuery(WorkspaceService.method.listPublicWorkspaces, {}, { select: (res) => res.workspaces });

/** 公開ワークスペースに参加する */
export const useJoinPublicWorkspace = () => {
  const queryClient = useQueryClient();

  return useMutation(WorkspaceService.method.joinPublicWorkspace, {
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: workspaceServiceKey });
    },
  });
};
