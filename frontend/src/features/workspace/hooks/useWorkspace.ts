import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

export const useWorkspaces = () => 
  useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data, error } = await api.GET("/api/workspaces", {});

      if (error || !data) {
        throw new Error(error?.error ?? "ワークスペースの取得に失敗しました");
      }

      return data.workspaces;
    },
  })
;

export const useCreateWorkspace = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: string; name: string; description?: string }) => {
      const { data: response, error } = await api.POST("/api/workspaces", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error ?? "ワークスペースの作成に失敗しました");
      }
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
};
