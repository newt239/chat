import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { toastError } from "#/lib/toastError";

/** ユーザーグループのメンバー一覧を取得する */
export const useUserGroupMembers = (groupId: string) =>
  useQuery(
    UserGroupService.method.listUserGroupMembers,
    { groupId },
    { select: (res) => res.members },
  );

/** ユーザーグループのメンバー追加・削除を提供する */
export const useUserGroupMemberActions = () => {
  const queryClient = useQueryClient();
  const onSuccess = async (_: object, { groupId }: { groupId?: string }) => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        input: { groupId },
        schema: UserGroupService.method.listUserGroupMembers,
      }),
    });
  };

  const add = useMutation(UserGroupService.method.addUserGroupMember, {
    onError: toastError,
    onSuccess,
  });
  const remove = useMutation(UserGroupService.method.removeUserGroupMember, {
    onError: toastError,
    onSuccess,
  });

  return { add, remove };
};
