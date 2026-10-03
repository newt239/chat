import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

import type { DescMessage, DescMethodUnary } from "@bufbuild/protobuf";

const workspaceServiceKey = createConnectQueryKey({
  cardinality: "finite",
  schema: WorkspaceService,
});

export const useWorkspaces = () =>
  useQuery(WorkspaceService.method.listWorkspaces, {}, { select: (res) => res.workspaces });

// ワークスペースの作成・更新・削除・参加。成功したらワークスペースの問い合わせをまとめて取り直す
export const useWorkspaceMutation = <I extends DescMessage, O extends DescMessage>(
  method: DescMethodUnary<I, O>,
) => {
  const queryClient = useQueryClient();
  return useMutation(method, {
    onSuccess: () => queryClient.invalidateQueries({ queryKey: workspaceServiceKey }),
  });
};
