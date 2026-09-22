import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

export const useWorkspaces = () =>
  useQuery({
    queryFn: async () => {
      const { data, error } = await api.GET("/api/workspaces", {});

      if (error) {
        throw new Error(error.error);
      }

      return data.workspaces;
    },
    queryKey: ["workspaces"],
  });

export const useCreateWorkspace = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: string; name: string; description?: string }) => {
      const { data: response, error } = await api.POST("/api/workspaces", {
        body: data,
      });
      if (error) {
        throw new Error(error.error);
      }
      return response;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
};
