import { useCallback } from "react";

import { useSendMessage } from "../hooks/useMessage";
import { BaseMessageInput } from "./BaseMessageInput";

type MessageInputProps = {
  channelId: string | null;
};

export const MessageInput = ({ channelId }: MessageInputProps) => {
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
      placeholder="メッセージを入力..."
      isPending={sendMessage.isPending}
      error={sendMessage.isError ? sendMessage.error.message : undefined}
      channelId={channelId}
    />
  );
};
