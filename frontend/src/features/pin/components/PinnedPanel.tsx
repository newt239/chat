import { IconPin } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton";
import { MessageLinkCard } from "#/features/message/components/MessageLinkCard";
import { usePinnedMessages } from "#/features/pin/hooks/usePinnedMessages";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "#/providers/store/workspace";

type PinnedPanelProps = {
  channelId: string | null;
};

export const PinnedPanel = ({ channelId }: PinnedPanelProps) => {
  const { t } = useTranslation();
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const effectiveChannelId = channelId ?? currentChannelId;
  const { pins, isLoading, isError } = usePinnedMessages(effectiveChannelId);

  if (!workspaceId || !effectiveChannelId) {
    return null;
  }

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
      <div className="flex flex-col items-center gap-3 p-8 font-sans text-body text-muted">
        <IconPin aria-hidden className="size-10 text-subtle" />
        {t("pin.empty")}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-0.5 overflow-y-auto p-1.5">
      {pins.map((pin) => (
        <MessageLinkCard
          key={pin.message.id}
          message={pin.message}
          workspaceId={workspaceId}
          markedAt={pin.pinnedAt}
        />
      ))}
    </div>
  );
};
