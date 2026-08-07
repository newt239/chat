import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

type CreateChannelInput = {
  name: string;
  description?: string;
  isPrivate?: boolean;
};

export const useChannels = (workspaceId: string | null) => 
  useQuery({
    queryKey: ["workspaces", workspaceId, "channels"],
    queryFn: async (): Promise<components["schemas"]["Channel"][]> => {
      if (workspaceId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/workspaces/{id}/channels", {
        params: { path: { id: workspaceId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "チャンネル一覧の取得に失敗しました");
      }

      return data;
    },
    enabled: workspaceId !== null,
  })
;

export const useCreateChannel = (workspaceId: string | null) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateChannelInput) => {
      if (workspaceId === null) {
        throw new Error("ワークスペースが選択されていません");
      }

      const { data, error } = await api.POST("/api/workspaces/{id}/channels", {
        params: { path: { id: workspaceId } },
        body: {
          name: input.name,
          description: input.description,
          isPrivate: input.isPrivate ?? false,
        },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "チャンネルの作成に失敗しました");
      }

      return data;
    },
    onSuccess: () => {
      if (workspaceId === null) {
        return;
      }
      queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "channels"] });
    },
  });
};
