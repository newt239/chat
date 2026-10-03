import { IconPin } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { MessageLinkCard } from "#/features/message/components/MessageLinkCard";
import { usePinnedMessages } from "#/features/pin/hooks/usePinnedMessages";

type PinnedPanelProps = {
  workspaceId: string;
  channelId: string;
};

export const PinnedPanel = ({ workspaceId, channelId }: PinnedPanelProps) => {
  const { t } = useTranslation();
  const { pins, isLoading, isError } = usePinnedMessages(channelId);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-3">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-12 w-full" />
      </div>
    );
  }

  if (isError) {
    return <p className="m-0 p-4 font-sans text-body text-danger">{t("pin.loadFailed")}</p>;
  }

  if (pins.length === 0) {
    return (
      <EmptyState icon={<IconPin />} title={t("pin.empty")} description={t("pin.emptyHint")} />
    );
  }

  return (
    <div className="flex flex-col gap-0.5 overflow-y-auto p-1.5">
      {pins.map((message) => (
        <MessageLinkCard
          key={message.id}
          message={message}
          workspaceId={workspaceId}
          markedAt={message.pin?.pinnedAt}
        />
      ))}
    </div>
  );
};
