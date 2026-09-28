import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

/** ユーザーグループのメンバー一覧を取得する */
export const useUserGroupMembers = (groupId: string | null) =>
  useQuery({
    enabled: groupId !== null,
    queryFn: async () => {
      if (groupId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/user-groups/{id}/members", {
        params: { path: { id: groupId } },
      });
      if (error) {
        throw new Error(error.error);
      }
      return data.members ?? [];
    },
    queryKey: ["user-groups", groupId, "members"],
  });

/** ユーザーグループのメンバー追加・削除を提供する */
export const useUserGroupMemberActions = (groupId: string) => {
  const queryClient = useQueryClient();

  const invalidate = async () => {
    await queryClient.invalidateQueries({ queryKey: ["user-groups", groupId, "members"] });
  };

  const add = useMutation({
    mutationFn: async (userId: string) => {
      const { error } = await api.POST("/api/user-groups/{id}/members", {
        body: { userId },
        params: { path: { id: groupId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: async (userId: string) => {
      const { error } = await api.DELETE("/api/user-groups/{id}/members", {
        params: { path: { id: groupId }, query: { userId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  return { add, remove };
};
