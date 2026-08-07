import { useQuery } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

export const useMembers = (workspaceId: string | null) =>
  useQuery({
    queryKey: ["workspaces", workspaceId, "members"],
    queryFn: async (): Promise<components["schemas"]["MemberInfo"][]> => {
      if (workspaceId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/workspaces/{id}/members", {
        params: { path: { id: workspaceId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "メンバー一覧の取得に失敗しました");
      }

      return data.members;
    },
    enabled: workspaceId !== null,
  });
