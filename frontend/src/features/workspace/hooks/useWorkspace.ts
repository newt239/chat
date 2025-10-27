import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

import type { WorkspaceSummary } from "@/features/workspace/types";

import { api } from "@/lib/api/client";

// 実際のAPIレスポンスの型定義
type WorkspacesResponse = {
  workspaces: WorkspaceSummary[];
};

export function useWorkspaces() {
  return useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data, error } = await api.GET("/api/workspaces", {});

      if (error || !data) {
        throw new Error(error?.error || "Failed to fetch workspaces");
      }

      // APIレスポンスは { workspaces: [...] } の形式なので、workspacesプロパティにアクセス
      // 型ガードを使用して安全に型チェック
      if (data && typeof data === "object" && "workspaces" in data) {
        const { workspaces: workspaceList } = data as WorkspacesResponse;
        return workspaceList || [];
      }

      // フォールバック: データが配列の場合はそのまま返す（後方互換性のため）
      if (Array.isArray(data)) {
        return data;
      }

      return [];
    },
  });
}

export function useCreateWorkspace() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { name: string; description?: string }) => {
      const { data: response, error } = await api.POST("/api/workspaces", {
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
