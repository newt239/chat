import { IconHash, IconLock } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton";
import { ChannelMemberManager } from "#/features/channel/components/ChannelMemberManager";
import { ChannelSettingsPanel } from "#/features/channel/components/ChannelSettingsPanel";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { currentChannelIdAtom } from "#/providers/store/workspace";

type ChannelInfoPanelProps = {
  workspaceId: string;
  channelId?: string | null;
};

export const ChannelInfoPanel = ({ workspaceId, channelId }: ChannelInfoPanelProps) => {
  const { t } = useTranslation();
  const { data: channels, isLoading, isError } = useChannels(workspaceId);
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const effectiveChannelId = channelId ?? currentChannelId;
  const activeChannel = channels?.find((candidate) => candidate.id === effectiveChannelId);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-4">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-4 w-48" />
      </div>
    );
  }

  if (isError || activeChannel === undefined) {
    return (
      <p className="m-0 p-4 text-caption text-muted">
        {isError ? t("channel.info.loadFailed") : t("channel.info.notFound")}
      </p>
    );
  }

  const description = activeChannel.description ?? "";

  return (
    <div className="flex min-h-full flex-col bg-surface font-sans text-text">
      <section className="flex flex-col gap-2 border-b border-border px-4 py-3">
        <h3 className="m-0 flex items-center gap-1.5 text-body-strong">
          {activeChannel.isPrivate ? (
            <IconLock aria-hidden className="size-4 text-muted" />
          ) : (
            <IconHash aria-hidden className="size-4 text-muted" />
          )}
          <span className="min-w-0 truncate">{activeChannel.name}</span>
        </h3>
        <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-3.5 gap-y-1 text-[12.5px]">
          <dt className="text-muted">{t("channel.info.visibility")}</dt>
          <dd className="m-0">
            {activeChannel.isPrivate ? t("channel.info.private") : t("channel.info.public")}
          </dd>
          <dt className="text-muted">{t("channel.info.id")}</dt>
          <dd className="m-0 truncate font-mono text-xs">{activeChannel.id}</dd>
        </dl>
      </section>
      <section className="flex flex-col gap-2 border-b border-border px-4 py-3">
        <h4 className="m-0 text-xs font-semibold text-muted">{t("channel.info.description")}</h4>
        {description.length > 0 ? (
          <p className="m-0 text-[13.5px] whitespace-pre-wrap">{description}</p>
        ) : (
          <p className="m-0 text-[12.5px] text-muted">{t("channel.info.noDescription")}</p>
        )}
      </section>
      <ChannelMemberManager channelId={activeChannel.id} workspaceId={workspaceId} />
      <ChannelSettingsPanel
        key={activeChannel.id}
        channelId={activeChannel.id}
        initialName={activeChannel.name}
        initialDescription={activeChannel.description ?? null}
        initialIsPrivate={activeChannel.isPrivate}
      />
    </div>
  );
};
