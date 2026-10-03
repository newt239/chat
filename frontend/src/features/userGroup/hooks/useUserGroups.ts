import { createConnectQueryKey, skipToken, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { toastError } from "#/lib/toastError";

/** ワークスペースのユーザーグループ一覧を取得する */
export const useUserGroups = (workspaceId: string | null) =>
  useQuery(
    UserGroupService.method.listUserGroups,
    workspaceId === null ? skipToken : { workspaceId },
    { select: (res) => res.userGroups },
  );

/** ユーザーグループの作成・更新・削除を提供する */
export const useUserGroupActions = () => {
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        schema: UserGroupService.method.listUserGroups,
      }),
    });
  };

  const create = useMutation(UserGroupService.method.createUserGroup, { onSuccess });
  const update = useMutation(UserGroupService.method.updateUserGroup, { onSuccess });
  const remove = useMutation(UserGroupService.method.deleteUserGroup, {
    onError: toastError,
    onSuccess,
  });

  return { create, remove, update };
};
