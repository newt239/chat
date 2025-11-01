import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { CreateDMRequest } from "../schemas";

import { api } from "@/lib/api/client";

export const useDMs = (workspaceId: string) => {
  return useQuery({
    queryKey: ["dms", workspaceId],
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
    enabled: !!workspaceId,
  });
};

export const useCreateDM = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateDMRequest) => {
      const response = await api.POST("/api/workspaces/{id}/dms", {
        params: {
          path: { id: workspaceId },
        },
        body: data,
      });

      if (response.error) {
        throw new Error(response.error.error || "DMの作成に失敗しました");
      }

      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dms", workspaceId] });
    },
  });
};
