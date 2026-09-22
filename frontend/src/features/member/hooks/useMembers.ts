import { useQuery } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

export const useMembers = (workspaceId: string | null) =>
  useQuery({
    enabled: workspaceId !== null,
    queryFn: async (): Promise<components["schemas"]["MemberInfo"][]> => {
      if (workspaceId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/workspaces/{id}/members", {
        params: { path: { id: workspaceId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data.members;
    },
    queryKey: ["workspaces", workspaceId, "members"],
  });
