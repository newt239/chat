import { useQuery } from "@tanstack/react-query";

import type { components } from "@/lib/api/schema";

import { participatingThreadsResponseSchema } from "@/features/thread/schemas";
import { api } from "@/lib/api/client";

type ParticipatingThreadsOutput = components["schemas"]["ParticipatingThreadsOutput"];

type UseParticipatingThreadsParams = {
  workspaceId: string | null;
  cursorLastActivityAt?: string;
  cursorThreadId?: string;
  limit?: number;
};

export function useParticipatingThreads(params: UseParticipatingThreadsParams) {
  const { workspaceId, cursorLastActivityAt, cursorThreadId, limit = 20 } = params;

  return useQuery({
    queryKey: [
      "participating-threads",
      workspaceId,
      cursorLastActivityAt ?? null,
      cursorThreadId ?? null,
      limit,
    ],
    enabled: typeof workspaceId === "string" && workspaceId.length > 0,
    queryFn: async (): Promise<ParticipatingThreadsOutput> => {
      if (!workspaceId) {
        return { items: [], next_cursor: undefined };
      }

      const { data, error } = await api.GET("/api/workspaces/{workspaceId}/threads/participating", {
        params: {
          path: { workspaceId },
          query: {
            cursorLastActivityAt,
            cursorThreadId,
            limit,
          },
        },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "参加中スレッドの取得に失敗しました");
      }

      const parsed = participatingThreadsResponseSchema.safeParse(data);
      if (!parsed.success) {
        throw new Error("スレッド一覧レスポンスの形式が想定と異なります");
      }

      return parsed.data;
    },
    staleTime: 15_000,
  });
}
