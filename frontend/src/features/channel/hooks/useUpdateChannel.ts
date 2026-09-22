import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

type UpdateChannelInput = {
  channelId: string;
  name?: string;
  description?: string;
  isPrivate?: boolean;
};

export const useUpdateChannel = (workspaceId: string | null) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ channelId, ...body }: UpdateChannelInput) => {
      const { data, error } = await api.PATCH("/api/channels/{channelId}", {
        body,
        params: { path: { channelId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: async () => {
      if (workspaceId !== null) {
        await queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "channels"] });
      }
    },
  });
};
