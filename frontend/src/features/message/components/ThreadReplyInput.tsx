import { useCallback } from "react";

import { BaseMessageInput } from "./BaseMessageInput";

type ThreadReplyInputProps = {
  channelId: string;
  onSubmit: (body: string) => void;
  isPending: boolean;
  isError: boolean;
  errorMessage?: string;
};

export const ThreadReplyInput = ({
  channelId,
  onSubmit,
  isPending,
  isError,
  errorMessage,
}: ThreadReplyInputProps) => {
  const handleSubmit = useCallback(
    (body: string) => {
      onSubmit(body);
    },
    [onSubmit]
  );

  return (
    <div className="border-t pt-4">
      <BaseMessageInput
        channelId={channelId}
        onSubmit={handleSubmit}
        placeholder="スレッドに返信..."
        isPending={isPending}
        error={isError ? (errorMessage ?? "返信の送信に失敗しました") : undefined}
      />
    </div>
  );
};
