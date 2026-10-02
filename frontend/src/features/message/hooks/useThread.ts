import { useEffect, useRef, useState } from "react";

import { callUnaryMethod, useMutation, useQuery } from "@connectrpc/connect-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";
import { transport } from "#/lib/api/transport";
import { useWsClient } from "#/providers/ws/useWsClient";

import { useInvalidateMessages } from "./useMessage";

import type { Message } from "#/gen/chat/v1/message_pb";

const THREAD_REPLIES_PAGE_SIZE = 50;

type Direction = "older" | "newer";

const appendNew = (current: Message[], added: Message[]) => [
  ...current,
  ...added.filter((message) => !current.some((reply) => reply.id === message.id)),
];

type ThreadState = {
  replies: Message[];
  replyCount: number;
  hasMore: boolean;
  hasNewer: boolean;
};

/** スレッドの返信を古い順に取得し、スクロールに合わせて前後を足す。aroundReplyId を渡すとその返信の前後から読む */
export const useThreadReplies = (threadId: string, aroundReplyId: string | null) => {
  const base = useQuery(ThreadService.method.getThreadReplies, {
    aroundReplyId: aroundReplyId ?? undefined,
    limit: THREAD_REPLIES_PAGE_SIZE,
    messageId: threadId,
  });
  const { wsClient } = useWsClient();
  // 取得し直すまでは前のスレッドの返信を出さないよう null にする
  const [state, setState] = useState<ThreadState | null>(null);
  const [loading, setLoading] = useState<Direction | null>(null);
  // スクロールのたびに呼ばれるため、読み込み中は重ねて取得しない
  const loadingRef = useRef(false);
  // 取り直す前の応答を捨てる
  const generationRef = useRef(0);

  useEffect(() => {
    generationRef.current += 1;
    loadingRef.current = false;
    setLoading(null);
    setState(
      base.data === undefined
        ? null
        : {
            hasMore: base.data.hasMore,
            hasNewer: base.data.hasNewer,
            replies: base.data.replies,
            replyCount: base.data.replyCount,
          },
    );
  }, [base.data]);

  const updateReplies = (update: (replies: Message[]) => Message[]) => {
    setState((prev) => prev && { ...prev, replies: update(prev.replies) });
  };

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }
    const unsubscribes = [
      wsClient.on("newMessage", ({ message }) => {
        if (message?.parentId !== threadId) {
          return;
        }
        // 最新まで読み込んでいるときだけ末尾に足す
        setState(
          (prev) =>
            prev && {
              ...prev,
              replies: prev.hasNewer ? prev.replies : appendNew(prev.replies, [message]),
              replyCount: prev.replyCount + 1,
            },
        );
      }),
      wsClient.on("messageUpdated", ({ message }) => {
        if (message?.parentId === threadId) {
          updateReplies((replies) =>
            replies.map((reply) => (reply.id === message.id ? message : reply)),
          );
        }
      }),
      wsClient.on("messageDeleted", ({ deletedMessageIds, deletedAt }) => {
        const deletedIds = new Set(deletedMessageIds);
        updateReplies((replies) =>
          replies.map((reply) =>
            deletedIds.has(reply.id) ? { ...reply, deletedAt, isDeleted: true } : reply,
          ),
        );
      }),
    ];
    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [wsClient, threadId]);

  const load = async (direction: Direction) => {
    const replies = state?.replies ?? [];
    const boundary = direction === "older" ? replies.at(0)?.createdAt : replies.at(-1)?.createdAt;
    if (boundary === undefined || loadingRef.current) {
      return;
    }
    const generation = generationRef.current;
    loadingRef.current = true;
    setLoading(direction);
    try {
      const response = await callUnaryMethod(transport, ThreadService.method.getThreadReplies, {
        limit: THREAD_REPLIES_PAGE_SIZE,
        messageId: threadId,
        ...(direction === "older" ? { until: boundary } : { since: boundary }),
      });
      if (generation !== generationRef.current) {
        return;
      }
      setState(
        (prev) =>
          prev &&
          (direction === "older"
            ? {
                ...prev,
                hasMore: response.hasMore,
                replies: appendNew(response.replies, prev.replies),
              }
            : {
                ...prev,
                hasNewer: response.hasNewer,
                replies: appendNew(prev.replies, response.replies),
              }),
      );
    } finally {
      if (generation === generationRef.current) {
        loadingRef.current = false;
        setLoading(null);
      }
    }
  };

  return {
    error: base.error,
    isError: base.isError,
    isLoading: base.isLoading || (base.data !== undefined && state === null),
    load,
    loading,
    parentMessage: base.data?.parentMessage,
    thread: state,
  };
};

/** スレッドに返信を送信するフック。返信は WebSocket で届くため、プレビューのためにメッセージ一覧だけ取り直す */
export const useSendThreadReply = () =>
  useMutation(MessageService.method.createMessage, { onSuccess: useInvalidateMessages() });
