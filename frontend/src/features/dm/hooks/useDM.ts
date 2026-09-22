import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { CreateDMRequest } from "../schemas";

export const useDMs = (workspaceId: string) =>
  useQuery({
    enabled: Boolean(workspaceId),
    queryFn: async () => {
      const response = await api.GET("/api/workspaces/{id}/dms", {
        params: {
          path: { id: workspaceId },
        },
      });

      if (response.error) {
        throw new Error(response.error.error || "DMの取得に失敗しました");
      }

      return response.data;
    },
    queryKey: ["dms", workspaceId],
  });

export const useCreateDM = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateDMRequest) => {
      const response = await api.POST("/api/workspaces/{id}/dms", {
        body: data,
        params: {
          path: { id: workspaceId },
        },
      });

      if (response.error) {
        throw new Error(response.error.error || "DMの作成に失敗しました");
      }

      return response.data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["dms", workspaceId] });
    },
  });
};
