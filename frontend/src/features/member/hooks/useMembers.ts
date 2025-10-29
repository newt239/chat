import { useQuery } from "@tanstack/react-query";

import type { components } from "@/lib/api/schema";

import { api } from "@/lib/api/client";

export function useMembers(workspaceId: string | null) {
  return useQuery({
    queryKey: ["workspaces", workspaceId, "members"],
    queryFn: async (): Promise<components["schemas"]["MemberInfo"][]> => {
      if (workspaceId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/workspaces/{id}/members", {
        params: { path: { id: workspaceId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "Failed to fetch members");
      }

      return data.members;
    },
    enabled: workspaceId !== null,
  });
}
