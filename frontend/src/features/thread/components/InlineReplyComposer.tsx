import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { BaseMessageInput } from "#/features/message/components/BaseMessageInput";
import { useInvalidateMessages } from "#/features/message/hooks/useMessage";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

import type { Message } from "#/gen/chat/v1/message_pb";

type InlineReplyComposerProps = {
  channelId: string;
  parentId: string;
  placeholder: string;
  onSent: (reply: Message) => void;
};

// 一覧のカードからその場でスレッドに返信する。送った返信は onSent で呼び出し側の表示に足す
export const InlineReplyComposer = ({
  channelId,
  parentId,
  placeholder,
  onSent,
}: InlineReplyComposerProps) => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const invalidateMessages = useInvalidateMessages();
  const send = useMutation(MessageService.method.createMessage, {
    onSuccess: async ({ message }) => {
      if (message) {
        onSent(message);
      }
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            cardinality: "finite",
            input: { messageId: parentId },
            schema: ThreadService.method.getThreadReplies,
          }),
        }),
        invalidateMessages(),
      ]);
    },
  });

  return (
    <BaseMessageInput
      channelId={channelId}
      placeholder={placeholder}
      isPending={send.isPending}
      error={send.isError ? t("message.thread.sendFailed") : undefined}
      onSubmit={(content) => {
        send.mutate({ ...content, channelId, parentId });
      }}
    />
  );
};
