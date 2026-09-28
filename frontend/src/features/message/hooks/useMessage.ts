import { createConnectQueryKey, skipToken, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";

export const MESSAGES_PAGE_SIZE = 50;

export const useMessages = (channelId: string | null) =>
  useQuery(
    MessageService.method.listMessages,
    channelId === null ? skipToken : { channelId, limit: MESSAGES_PAGE_SIZE },
  );

/** スレッド付きの一覧も含め、メッセージ一覧を再取得する */
export const useInvalidateMessages = () => {
  const queryClient = useQueryClient();

  return async () => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({ cardinality: "finite", schema: MessageService }),
    });
  };
};

export const useSendMessage = () =>
  useMutation(MessageService.method.createMessage, { onSuccess: useInvalidateMessages() });

export const useUpdateMessage = () =>
  useMutation(MessageService.method.updateMessage, { onSuccess: useInvalidateMessages() });

export const useDeleteMessage = () =>
  useMutation(MessageService.method.deleteMessage, { onSuccess: useInvalidateMessages() });
