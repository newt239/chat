import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

import type { components } from "@/lib/api/schema";

import { apiClient } from "@/lib/api/client";

export function useWorkspaces() {
  return useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data, error } = await apiClient.GET("/api/workspaces", {});
      if (error || !data) {
        throw new Error(error?.error || "Failed to fetch workspaces");
      }
      return data;
    },
  });
}

export function useWorkspace(workspaceId: string) {
  return useQuery({
    queryKey: ["workspaces", workspaceId],
    queryFn: async (): Promise<components["schemas"]["Workspace"] | undefined> => {
      // Note: This endpoint doesn't exist in the API schema
      // This function may need to be implemented differently or removed
      const { data } = await apiClient.GET("/api/workspaces", {});
      if (!data) {
        throw new Error("Failed to fetch workspaces");
      }
      return data.find((workspace) => workspace.id === workspaceId);
    },
    enabled: !!workspaceId,
  });
}

export function useCreateWorkspace() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { name: string; description?: string }) => {
      const { data: response, error } = await apiClient.POST("/api/workspaces", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error || "Failed to create workspace");
      }
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}
