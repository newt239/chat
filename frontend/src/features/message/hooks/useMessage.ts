import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";

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
  const invalidateThreadMetadata = useInvalidateThreadMetadata();
  return useMutation(MessageService.method.createMessage, {
    // 返信のときだけ親のスレッドの件数が変わる
    onSuccess: async ({ message }) => {
      if (message?.parentId !== undefined) {
        await invalidateThreadMetadata(message.channelId);
      }
    },
  });
};

export const useUpdateMessage = () => useMutation(MessageService.method.updateMessage);

export const useDeleteMessage = () => useMutation(MessageService.method.deleteMessage);
