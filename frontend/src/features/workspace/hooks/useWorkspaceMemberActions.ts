import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

type WorkspaceRole = components["schemas"]["UpdateMemberRoleRequest"]["role"];

/** ワークスペースメンバーの招待・削除・ロール変更をまとめて提供する */
export const useWorkspaceMemberActions = (workspaceId: string) => {
  const queryClient = useQueryClient();

  const invalidate = async () => {
    await queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "members"] });
  };

  const invite = useMutation({
    mutationFn: async ({ email, role }: { email: string; role: WorkspaceRole }) => {
      const { error } = await api.POST("/api/workspaces/{id}/members", {
        body: { email, role },
        params: { path: { id: workspaceId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: async ({ userId }: { userId: string }) => {
      const { error } = await api.DELETE("/api/workspaces/{id}/members/{userId}", {
        params: { path: { id: workspaceId, userId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const updateRole = useMutation({
    mutationFn: async ({ userId, role }: { userId: string; role: WorkspaceRole }) => {
      const { error } = await api.PATCH("/api/workspaces/{id}/members/{userId}", {
        body: { role },
        params: { path: { id: workspaceId, userId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  return { invite, remove, updateRole };
};
