import { callUnaryMethod, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useInfiniteQuery } from "@tanstack/react-query";

import { MentionService } from "#/gen/chat/v1/mention_service_pb";

import type { MentionCursor } from "#/gen/chat/v1/mention_service_pb";

// connect-query の useInfiniteQuery は最初のカーソルに undefined を渡せないため、TanStack Query を直接使う
export const useMentions = (workspaceId: string) => {
  const transport = useTransport();
  return useInfiniteQuery({
    getNextPageParam: (res) => res.nextCursor,
    initialPageParam: undefined,
    queryFn: ({
      pageParam,
      signal,
    }: {
      pageParam: MentionCursor | undefined;
      signal: AbortSignal;
    }) =>
      callUnaryMethod(
        transport,
        MentionService.method.listMentions,
        { cursor: pageParam, limit: 20, workspaceId },
        { signal },
      ),
    queryKey: createConnectQueryKey({
      cardinality: "infinite",
      input: { workspaceId },
      schema: MentionService.method.listMentions,
      transport,
    }),
    select: (data) => data.pages.flatMap((page) => page.messages),
  });
};
