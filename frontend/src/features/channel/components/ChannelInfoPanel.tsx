import { IconHash, IconLock } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { LinkButton } from "#/components/ui/LinkButton/LinkButton";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { surfaceNavTone } from "#/components/ui/styles/navTone";
import { Switch } from "#/components/ui/Switch/Switch";
import { ChannelAppsSection } from "#/features/app/components/ChannelAppsSection";
import { ChannelMemberManager } from "#/features/channel/components/ChannelMemberManager";
import { ChannelSettingsPanel } from "#/features/channel/components/ChannelSettingsPanel";
import { canHaveChildChannel } from "#/features/channel/utils/channelPath";
import { openDialog } from "#/lib/overlaySearch";

import { useChannelAggregation } from "../hooks/useChannelAggregation";
import { useChannelListActions } from "../hooks/useChannelListActions";
import { relativePath } from "../utils/channelTree";
import { ChannelLinksSection } from "./ChannelLinksSection";
import { ChannelNavItem } from "./ChannelNavItem";

type ChannelInfoPanelProps = {
  workspaceId: string;
  channelId: string;
};

export const ChannelInfoPanel = ({ workspaceId, channelId }: ChannelInfoPanelProps) => {
  const { t } = useTranslation();
  const {
    channel: activeChannel,
    descendants,
    isError,
    isResolved,
  } = useChannelAggregation(workspaceId, channelId);
  const { setMuted, setStarred } = useChannelListActions(workspaceId);

  if (!isResolved) {
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
  const canCreateChild = activeChannel.isMember && canHaveChildChannel(activeChannel.name);

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
        <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-3.5 gap-y-1 text-label font-normal">
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
          <p className="m-0 text-body-sm whitespace-pre-wrap">{description}</p>
        ) : (
          <p className="m-0 text-label font-normal text-muted">{t("channel.info.noDescription")}</p>
        )}
      </section>
      <ChannelLinksSection channelId={activeChannel.id} />
      {(descendants.length > 0 || canCreateChild) && (
        <section
          className={`flex flex-col gap-1.5 border-b border-border px-4 py-3 ${surfaceNavTone}`}
        >
          <h4 className="m-0 flex items-center justify-between text-xs font-semibold text-muted">
            {t("channel.info.descendants", { count: descendants.length })}
            {canCreateChild && (
              <LinkButton
                size="sm"
                variant="ghost"
                to="."
                search={openDialog({ dialog: "create-channel", parent: activeChannel.id })}
              >
                {t("channel.info.createChild")}
              </LinkButton>
            )}
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
          {descendants.length > 0 && (
            <p className="m-0 text-caption text-subtle">{t("channel.info.descendantsHint")}</p>
          )}
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
      <ChannelAppsSection workspaceId={workspaceId} channelId={activeChannel.id} />
      <ChannelSettingsPanel
        key={activeChannel.id}
        workspaceId={workspaceId}
        channel={activeChannel}
      />
    </div>
  );
};
