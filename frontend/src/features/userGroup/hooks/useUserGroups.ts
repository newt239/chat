import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

type CreateUserGroupInput = {
  name: string;
  description?: string;
};

/** ワークスペースのユーザーグループ一覧を取得する */
export const useUserGroups = (workspaceId: string | null) =>
  useQuery({
    enabled: workspaceId !== null,
    queryFn: async () => {
      if (workspaceId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/user-groups", {
        params: { query: { workspaceId } },
      });
      if (error) {
        throw new Error(error.error);
      }
      return data.userGroups ?? [];
    },
    queryKey: ["workspaces", workspaceId, "user-groups"],
  });

/** ユーザーグループの作成・更新・削除を提供する */
export const useUserGroupActions = (workspaceId: string) => {
  const queryClient = useQueryClient();

  const invalidate = async () => {
    await queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "user-groups"] });
  };

  const create = useMutation({
    mutationFn: async (input: CreateUserGroupInput) => {
      const { data, error } = await api.POST("/api/user-groups", {
        body: { ...input, workspaceId },
      });
      if (error) {
        throw new Error(error.error);
      }
      return data;
    },
    onSuccess: invalidate,
  });

  const update = useMutation({
    mutationFn: async ({ id, ...body }: CreateUserGroupInput & { id: string }) => {
      const { error } = await api.PATCH("/api/user-groups/{id}", {
        body,
        params: { path: { id } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE("/api/user-groups/{id}", {
        params: { path: { id } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  return { create, remove, update };
};
