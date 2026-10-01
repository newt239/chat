import { IconHash, IconLock } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { Switch } from "#/components/ui/Switch/Switch";
import { ChannelMemberManager } from "#/features/channel/components/ChannelMemberManager";
import { ChannelSettingsPanel } from "#/features/channel/components/ChannelSettingsPanel";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { WebhooksSection } from "#/features/webhook/components/WebhooksSection";

import { useChannelListActions } from "../hooks/useChannelListActions";
import { isDescendantPath, relativePath } from "../utils/channelTree";
import { ChannelLinksSection } from "./ChannelLinksSection";
import { ChannelNavItem } from "./ChannelNavItem";

type ChannelInfoPanelProps = {
  workspaceId: string;
  channelId: string;
};

export const ChannelInfoPanel = ({ workspaceId, channelId }: ChannelInfoPanelProps) => {
  const { t } = useTranslation();
  const { data: channels, isLoading, isError } = useChannels(workspaceId);
  const activeChannel = channels?.find((candidate) => candidate.id === channelId);
  const { setMuted, setStarred } = useChannelListActions(workspaceId);

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
  const descendants = (channels ?? []).filter((candidate) =>
    isDescendantPath(activeChannel.name, candidate.name),
  );

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
      <ChannelLinksSection channelId={activeChannel.id} />
      {descendants.length > 0 && (
        <section className="flex flex-col gap-1.5 border-b border-border px-4 py-3 [--nav-active-fg:var(--c-accent-text)] [--nav-active:var(--c-accent-soft)] [--nav-fg:var(--c-text)] [--nav-hover:var(--c-hover)] [--nav-muted:var(--c-muted)] [--nav-strong:var(--c-text)]">
          <h4 className="m-0 text-xs font-semibold text-muted">
            {t("channel.info.descendants", { count: descendants.length })}
          </h4>
          <div className="-mx-2 flex flex-col">
            {descendants.map((descendant) => (
              <ChannelNavItem
                key={descendant.id}
                workspaceId={workspaceId}
                channelId={descendant.id}
                isStarred={descendant.isStarred}
                isMuted={descendant.isMuted}
                unreadCount={descendant.unreadCount}
                badgeCount={descendant.mentionCount}
              >
                {descendant.isPrivate ? <IconLock aria-hidden /> : <IconHash aria-hidden />}
                <span className="min-w-0 flex-1 truncate">
                  {relativePath(activeChannel.name, descendant.name)}
                </span>
              </ChannelNavItem>
            ))}
          </div>
          <p className="m-0 text-[11.5px] text-subtle">{t("channel.info.descendantsHint")}</p>
        </section>
      )}
      <section className="flex flex-col gap-3 border-b border-border px-4 py-3">
        <Switch
          isSelected={activeChannel.isStarred}
          onChange={(isSelected) => {
            setStarred(activeChannel.id, isSelected);
          }}
        >
          {t("channel.info.star")}
        </Switch>
        <Switch
          isSelected={activeChannel.isMuted}
          onChange={(isSelected) => {
            setMuted(activeChannel.id, isSelected);
          }}
        >
          {t("channel.info.mute")}
        </Switch>
      </section>
      <ChannelMemberManager channelId={activeChannel.id} workspaceId={workspaceId} />
      <WebhooksSection channelId={activeChannel.id} />
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
