import { BaseMessageInput } from "./BaseMessageInput";

type ThreadReplyInputProps = {
  channelId: string;
  onSubmit: (body: string, attachmentIds: string[]) => void;
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
}: ThreadReplyInputProps) => (
  <div className="border-t pt-4">
    <BaseMessageInput
      channelId={channelId}
      onSubmit={onSubmit}
      placeholder="スレッドに返信..."
      isPending={isPending}
      error={isError ? (errorMessage ?? "返信の送信に失敗しました") : undefined}
    />
  </div>
);
