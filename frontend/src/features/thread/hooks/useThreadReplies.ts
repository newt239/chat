import { useEffect } from "react";

import { callUnaryMethod, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { toRange, useBidirectionalPages } from "#/features/message/hooks/useBidirectionalPages";
import { subscribeMessagePatches } from "#/features/message/utils/updateTimelineMessage";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";
import { useWsClient } from "#/providers/ws/useWsClient";

import type { PageCursor } from "#/features/message/hooks/useBidirectionalPages";
import type { GetThreadRepliesResponse } from "#/gen/chat/v1/thread_service_pb";

import type { InfiniteData } from "@tanstack/react-query";

const THREAD_REPLIES_PAGE_SIZE = 50;

const getReplies = (page: GetThreadRepliesResponse) => page.replies;

/** スレッドの返信を古い順に取得し、スクロールに合わせて前後を足す。aroundReplyId を渡すとその返信の前後から読む */
export const useThreadReplies = (threadId: string, aroundReplyId: string | null) => {
  const transport = useTransport();
  const queryClient = useQueryClient();
  const wsClient = useWsClient();
  const input = {
    aroundReplyId: aroundReplyId ?? undefined,
    limit: THREAD_REPLIES_PAGE_SIZE,
    messageId: threadId,
  };
  const pages = useBidirectionalPages({
    enabled: true,
    fetchPage: (cursor: PageCursor, signal) =>
      callUnaryMethod(
        transport,
        ThreadService.method.getThreadReplies,
        cursor === null ? input : { ...input, aroundReplyId: undefined, ...toRange(cursor) },
        { signal },
      ),
    getItems: getReplies,
    newestFirst: false,
    queryKey: createConnectQueryKey({
      cardinality: "infinite",
      input,
      schema: ThreadService.method.getThreadReplies,
      transport,
    }),
  });

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }
    const updatePages = (
      update: (page: GetThreadRepliesResponse, isLatestPage: boolean) => GetThreadRepliesResponse,
    ) => {
      queryClient.setQueriesData<InfiniteData<GetThreadRepliesResponse>>(
        {
          queryKey: createConnectQueryKey({
            cardinality: "infinite",
            input: { messageId: threadId },
            schema: ThreadService.method.getThreadReplies,
          }),
        },
        (data) =>
          data && {
            ...data,
            pages: data.pages.map((page, index) =>
              update(page, index === data.pages.length - 1 && !page.hasNewer),
            ),
          },
      );
    };
    const unsubscribes = [
      wsClient.on("newMessage", ({ message }) => {
        if (message?.parentId !== threadId) {
          return;
        }
        // 古い順に並ぶので末尾のページが最新。最新まで読み込んでいるときだけ足す
        updatePages((page, isLatestPage) => ({
          ...page,
          replies:
            isLatestPage && !page.replies.some((reply) => reply.id === message.id)
              ? [...page.replies, message]
              : page.replies,
          replyCount: page.replyCount + 1,
        }));
      }),
      // 親メッセージはパネルの先頭に出るので返信と同じく更新する
      subscribeMessagePatches(wsClient, (_channelId, messageIds, update) => {
        updatePages((page) => ({
          ...page,
          parentMessage:
            page.parentMessage && messageIds.has(page.parentMessage.id)
              ? update(page.parentMessage)
              : page.parentMessage,
          replies: page.replies.map((reply) => (messageIds.has(reply.id) ? update(reply) : reply)),
        }));
      }),
    ];
    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [wsClient, queryClient, threadId]);

  const latestPage = pages.pages?.at(-1);
  return {
    ...pages,
    parentMessage: latestPage?.parentMessage,
    replyCount: latestPage?.replyCount ?? 0,
  };
};
