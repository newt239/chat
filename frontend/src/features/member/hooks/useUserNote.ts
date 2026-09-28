import { createConnectQueryKey, skipToken, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { UserService } from "#/gen/chat/v1/user_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

export const useUserNote = (targetUserId: string | null) =>
  useQuery(UserService.method.getUserNote, targetUserId === null ? skipToken : { targetUserId }, {
    select: (res) => res.note,
  });

// ニックネームはメンバー一覧から引いて表示するため、メンバー一覧も取り直す
export const useUpdateUserNote = () => {
  const queryClient = useQueryClient();
  return useMutation(UserService.method.updateUserNote, {
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            cardinality: "finite",
            schema: UserService.method.getUserNote,
          }),
        }),
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            cardinality: "finite",
            schema: WorkspaceService.method.listMembers,
          }),
        }),
      ]);
    },
  });
};
