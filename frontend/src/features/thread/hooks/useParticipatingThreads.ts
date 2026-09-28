import { callUnaryMethod, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";

import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

import type {
  ListParticipatingThreadsResponse,
  ParticipatingThread,
  ThreadCursor,
} from "#/gen/chat/v1/thread_service_pb";

import type { InfiniteData } from "@tanstack/react-query";

const threadListKey = createConnectQueryKey({
  cardinality: "infinite",
  schema: ThreadService.method.listParticipatingThreads,
});

// connect-query の useInfiniteQuery は最初のカーソルに undefined を渡せないため、TanStack Query を直接使う
export const useParticipatingThreads = (workspaceId: string) => {
  const transport = useTransport();
  return useInfiniteQuery({
    getNextPageParam: (res) => res.nextCursor,
    initialPageParam: undefined,
    queryFn: ({
      pageParam,
      signal,
    }: {
      pageParam: ThreadCursor | undefined;
      signal: AbortSignal;
    }) =>
      callUnaryMethod(
        transport,
        ThreadService.method.listParticipatingThreads,
        { cursor: pageParam, limit: 20, workspaceId },
        { signal },
      ),
    queryKey: createConnectQueryKey({
      cardinality: "infinite",
      input: { workspaceId },
      schema: ThreadService.method.listParticipatingThreads,
      transport,
    }),
    // 最初のメッセージが削除されたスレッドは表示できないため除く
    select: (data) =>
      data.pages.flatMap((page) =>
        page.threads.flatMap(({ firstMessage, ...thread }) =>
          firstMessage === undefined ? [] : [{ ...thread, firstMessage }],
        ),
      ),
    staleTime: 15_000,
  });
};

// 一覧のキャッシュの 1 件を書き換える。再取得しないので、入力中のカードの並び順が変わらない
export const useUpdateListedThread = () => {
  const queryClient = useQueryClient();
  return (threadId: string, update: (thread: ParticipatingThread) => ParticipatingThread) => {
    queryClient.setQueriesData<InfiniteData<ListParticipatingThreadsResponse>>(
      { queryKey: threadListKey },
      (data) =>
        data === undefined
          ? data
          : {
              ...data,
              pages: data.pages.map((page) => ({
                ...page,
                threads: page.threads.map((thread) =>
                  thread.threadId === threadId ? update(thread) : thread,
                ),
              })),
            },
    );
  };
};
