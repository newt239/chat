import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";

export function useWorkspaces() {
  return useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data, error } = await apiClient.GET("/api/workspaces", {});
      if (error || !data) {
        throw new Error((error as any)?.error || "Failed to fetch workspaces");
      }
      return (data as any) || [];
    },
  });
}

export function useWorkspace(workspaceId: string) {
  return useQuery({
    queryKey: ["workspaces", workspaceId],
    queryFn: async () => {
      const { data, error } = await apiClient.GET("/api/workspaces/{id}" as any, {
        params: { path: { id: workspaceId } },
      });
      if (error || !data) {
        throw new Error((error as any)?.error || "Failed to fetch workspace");
      }
      return data as any;
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
        throw new Error((error as any)?.error || "Failed to create workspace");
      }
      return response as any;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}
