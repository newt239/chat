import { IconClock, IconSend } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { useConversationLabel } from "#/features/channel/hooks/useConversationLabel";
import { ScheduledMessageStatus } from "#/gen/chat/v1/scheduled_message_service_pb";

import { useScheduledMessages } from "../hooks/useScheduledMessages";
import { ScheduledMessageItem } from "./ScheduledMessageItem";

const isSent = (status: ScheduledMessageStatus) => status === ScheduledMessageStatus.SENT;

type ScheduledMessageListProps = {
  workspaceId: string;
  // pending は予約中・送信中・失敗を送信予定の順に、sent は送信済みを新しい順に並べる
  mode: "pending" | "sent";
};

export const ScheduledMessageList = ({ workspaceId, mode }: ScheduledMessageListProps) => {
  const { t } = useTranslation();
  const { data: messages, isLoading } = useScheduledMessages(workspaceId);
  const labelOf = useConversationLabel(workspaceId);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-3">
        <Skeleton className="h-14 w-full" />
        <Skeleton className="h-14 w-full" />
      </div>
    );
  }

  const items = (messages ?? []).filter((message) => isSent(message.status) === (mode === "sent"));
  if (mode === "sent") {
    items.reverse();
  }
  if (items.length === 0) {
    return mode === "sent" ? (
      <EmptyState
        icon={<IconSend />}
        title={t("schedule.list.sentEmpty")}
        description={t("schedule.list.sentEmptyHint")}
      />
    ) : (
      <EmptyState
        icon={<IconClock />}
        title={t("schedule.list.empty")}
        description={t("schedule.list.emptyHint")}
      />
    );
  }

  return (
    <div className="flex flex-col gap-2 p-3">
      {items.map((message) => (
        <ScheduledMessageItem
          key={message.id}
          workspaceId={workspaceId}
          message={message}
          label={labelOf(message.channelId)}
        />
      ))}
    </div>
  );
};
