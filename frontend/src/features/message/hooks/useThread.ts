import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

import { useInvalidateMessages } from "./useMessage";

/** スレッドの返信一覧を取得するフック */
export const useThreadReplies = (messageId: string) =>
  useQuery(ThreadService.method.getThreadReplies, { messageId });

/** スレッドに返信を送信するフック */
export const useSendThreadReply = () => {
  const queryClient = useQueryClient();
  const invalidateMessages = useInvalidateMessages();

  return useMutation(MessageService.method.createMessage, {
    onSuccess: async () => {
      // スレッドの返信とメタデータに加え、スレッドプレビューのためにメッセージ一覧も再取得する
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({ cardinality: "finite", schema: ThreadService }),
        }),
        invalidateMessages(),
      ]);
    },
  });
};
