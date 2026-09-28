import { useTranslation } from "react-i18next";

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
}: ThreadReplyInputProps) => {
  const { t } = useTranslation();

  return (
    <BaseMessageInput
      channelId={channelId}
      onSubmit={onSubmit}
      placeholder={t("message.thread.replyPlaceholder")}
      isPending={isPending}
      error={isError ? (errorMessage ?? t("message.thread.sendFailed")) : undefined}
    />
  );
};
