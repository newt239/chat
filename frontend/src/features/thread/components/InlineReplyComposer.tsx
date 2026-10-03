import { useTranslation } from "react-i18next";

import { BaseMessageInput } from "#/features/message/components/BaseMessageInput";
import { useSendMessage } from "#/features/message/hooks/useMessage";

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
  const send = useSendMessage();

  return (
    <BaseMessageInput
      channelId={channelId}
      parentId={parentId}
      placeholder={placeholder}
      isPending={send.isPending}
      error={send.isError ? t("message.thread.sendFailed") : null}
      targetPicker={null}
      onSubmit={(content) => {
        send.mutate(
          { ...content, channelId, parentId },
          {
            onSuccess: ({ message }) => {
              if (message) {
                onSent(message);
              }
            },
          },
        );
      }}
    />
  );
};
