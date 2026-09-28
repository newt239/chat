import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

type UpdateWorkspaceInput = components["schemas"]["UpdateWorkspaceRequest"];

/** ワークスペース自体の更新・削除を提供する */
export const useWorkspaceActions = (workspaceId: string) => {
  const queryClient = useQueryClient();

  const update = useMutation({
    mutationFn: async (input: UpdateWorkspaceInput) => {
      const { error } = await api.PATCH("/api/workspaces/{id}", {
        body: input,
        params: { path: { id: workspaceId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });

  const remove = useMutation({
    mutationFn: async () => {
      const { error } = await api.DELETE("/api/workspaces/{id}", {
        params: { path: { id: workspaceId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });

  return { remove, update };
};
