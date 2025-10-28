import { useCallback } from "react";

import { useSendMessage } from "../hooks/useMessage";

import { BaseMessageInput } from "./BaseMessageInput";

type MessageInputProps = {
  channelId: string | null;
};

export const MessageInput = ({ channelId }: MessageInputProps) => {
  const sendMessage = useSendMessage(channelId);

  const handleSubmit = useCallback(
    (body: string, attachmentIds: string[]) => {
      sendMessage.mutate({ body, attachmentIds });
    },
    [sendMessage]
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
      error={
        sendMessage.isError
          ? (sendMessage.error?.message ?? "メッセージの送信に失敗しました")
          : undefined
      }
      channelId={channelId}
    />
  );
};
