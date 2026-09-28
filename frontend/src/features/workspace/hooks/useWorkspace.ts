import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

export const workspaceServiceKey = createConnectQueryKey({
  cardinality: "finite",
  schema: WorkspaceService,
});

export const useWorkspaces = () =>
  useQuery(WorkspaceService.method.listWorkspaces, {}, { select: (res) => res.workspaces });

export const useCreateWorkspace = () => {
  const queryClient = useQueryClient();

  return useMutation(WorkspaceService.method.createWorkspace, {
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: workspaceServiceKey });
    },
  });
};
