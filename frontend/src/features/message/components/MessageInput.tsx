import { useCallback } from "react";

import { useTranslation } from "react-i18next";

import { useSendMessage } from "../hooks/useMessage";
import { BaseMessageInput } from "./BaseMessageInput";

type MessageInputProps = {
  channelId: string | null;
};

export const MessageInput = ({ channelId }: MessageInputProps) => {
  const { t } = useTranslation();
  const sendMessage = useSendMessage();

  const handleSubmit = useCallback(
    (body: string, attachmentIds: string[]) => {
      if (channelId !== null) {
        sendMessage.mutate({ attachmentIds, body, channelId });
      }
    },
    [sendMessage, channelId],
  );

  if (!channelId) {
    return null;
  }

  return (
    <BaseMessageInput
      key={channelId}
      onSubmit={handleSubmit}
      placeholder={t("message.composer.placeholder")}
      isPending={sendMessage.isPending}
      error={sendMessage.isError ? sendMessage.error.message : undefined}
      channelId={channelId}
    />
  );
};
