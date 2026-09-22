import { useQuery } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

type ParticipatingThreadsOutput = components["schemas"]["ParticipatingThreadsOutput"];

type UseParticipatingThreadsParams = {
  workspaceId: string | null;
  cursorLastActivityAt?: string;
  cursorThreadId?: string;
  limit?: number;
};

export const useParticipatingThreads = (params: UseParticipatingThreadsParams) => {
  const { workspaceId, cursorLastActivityAt, cursorThreadId, limit = 20 } = params;

  return useQuery({
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

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    queryKey: [
      "participating-threads",
      workspaceId,
      cursorLastActivityAt ?? null,
      cursorThreadId ?? null,
      limit,
    ],
    staleTime: 15_000,
  });
};
