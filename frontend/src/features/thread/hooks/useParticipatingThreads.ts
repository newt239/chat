import { skipToken, useQuery } from "@connectrpc/connect-query";

import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

import type { ThreadCursor } from "#/gen/chat/v1/thread_service_pb";

export const useParticipatingThreads = (workspaceId: string, cursor: ThreadCursor | undefined) =>
  useQuery(
    ThreadService.method.listParticipatingThreads,
    workspaceId.length > 0 ? { cursor, limit: 20, workspaceId } : skipToken,
    {
      // 最初のメッセージが削除されたスレッドは表示できないため除く
      select: (res) => ({
        nextCursor: res.nextCursor,
        threads: res.threads.flatMap(({ firstMessage, ...thread }) =>
          firstMessage === undefined ? [] : [{ ...thread, firstMessage }],
        ),
      }),
      staleTime: 15_000,
    },
  );
