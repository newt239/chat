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
    enabled: workspaceId !== null,
    queryFn: async (): Promise<components["schemas"]["Channel"][]> => {
      if (workspaceId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/workspaces/{id}/channels", {
        params: { path: { id: workspaceId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    queryKey: ["workspaces", workspaceId, "channels"],
  });

export const useCreateChannel = (workspaceId: string | null) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateChannelInput) => {
      if (workspaceId === null) {
        throw new Error("ワークスペースが選択されていません");
      }

      const { data, error } = await api.POST("/api/workspaces/{id}/channels", {
        body: {
          description: input.description,
          isPrivate: input.isPrivate ?? false,
          name: input.name,
        },
        params: { path: { id: workspaceId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: () => {
      if (workspaceId === null) {
        return;
      }
      void queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "channels"] });
    },
  });
};
