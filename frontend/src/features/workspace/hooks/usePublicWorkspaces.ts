import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

/** 参加できる公開ワークスペースの一覧を取得する */
export const usePublicWorkspaces = () =>
  useQuery({
    queryFn: async () => {
      const { data, error } = await api.GET("/api/workspaces/public");
      if (error) {
        throw new Error(error.error);
      }
      return data.workspaces;
    },
    queryKey: ["workspaces", "public"],
  });

/** 公開ワークスペースに参加する */
export const useJoinPublicWorkspace = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workspaceId: string) => {
      const { error } = await api.POST("/api/workspaces/{id}/join", {
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
};
