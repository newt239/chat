import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

/** チャンネルのスレッドの件数（メタデータ）を取り直す。本文の変化は WebSocket の差分でタイムラインに当てる */
export const useInvalidateThreadMetadata = () => {
  const queryClient = useQueryClient();
  return (channelId: string) =>
    queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        input: { channelId },
        schema: MessageService.method.listMessagesWithThread,
      }),
    });
};

export const useSendMessage = () => {
  const queryClient = useQueryClient();
  const invalidateThreadMetadata = useInvalidateThreadMetadata();
  return useMutation(MessageService.method.createMessage, {
    // 返信のときだけ親のスレッドの件数と返信の一覧が変わる
    onSuccess: async ({ message }) => {
      if (message?.parentId === undefined) {
        return;
      }
      await Promise.all([
        invalidateThreadMetadata(message.channelId),
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            cardinality: "infinite",
            input: { messageId: message.parentId },
            schema: ThreadService.method.getThreadReplies,
          }),
        }),
      ]);
    },
  });
};
