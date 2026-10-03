import {
  IconBell,
  IconBellOff,
  IconChecks,
  IconExternalLink,
  IconLink,
  IconSquarePlus,
  IconStar,
  IconStarOff,
} from "@tabler/icons-react";
import { useNavigate, useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { useChannelListActions } from "#/features/channel/hooks/useChannelListActions";
import { canHaveChildChannel } from "#/features/channel/utils/channelPath";
import { copyWithToast } from "#/lib/clipboard";
import { openDialog } from "#/lib/overlaySearch";
import { isTauri } from "#/lib/platform/platform";
import { toShareUrl } from "#/lib/shareUrl";

import { MoveToCategoryMenu } from "./MoveToCategoryMenu";

type ChannelMenuItemsProps = {
  workspaceId: string;
  channelId: string;
  isStarred: boolean;
  isMuted: boolean;
};

// サイドバーの右クリックとヘッダーの「その他」で共有するチャンネル・DM の操作
export const ChannelMenuItems = ({
  workspaceId,
  channelId,
  isStarred,
  isMuted,
}: ChannelMenuItemsProps) => {
  const { t } = useTranslation();
  const router = useRouter();
  const navigate = useNavigate();
  const { markAsRead, setMuted, setStarred } = useChannelListActions(workspaceId);
  const { data: channels = [] } = useChannels(workspaceId);
  // DM は一覧にないためカテゴリに入れられない
  const channel = channels.find((candidate) => candidate.id === channelId);
  const location = {
    params: { channelId, workspaceId },
    to: "/app/$workspaceId/$channelId",
  } as const;

  return (
    <>
      <MenuItem
        icon={isStarred ? <IconStarOff /> : <IconStar />}
        onAction={() => {
          setStarred(channelId, !isStarred);
        }}
      >
        {isStarred ? t("shell.channelMenu.unstar") : t("shell.channelMenu.star")}
      </MenuItem>
      <MenuItem
        icon={isMuted ? <IconBell /> : <IconBellOff />}
        onAction={() => {
          setMuted(channelId, !isMuted);
        }}
      >
        {isMuted ? t("shell.channelMenu.unmute") : t("shell.channelMenu.mute")}
      </MenuItem>
      {channel && (
        <MoveToCategoryMenu workspaceId={workspaceId} channel={channel} channels={channels} />
      )}
      {channel?.isMember && canHaveChildChannel(channel.name) && (
        <MenuItem
          icon={<IconSquarePlus />}
          onAction={() => {
            void navigate({
              search: openDialog({ dialog: "create-channel", parent: channelId }),
              to: ".",
            });
          }}
        >
          {t("shell.channelMenu.createChild")}
        </MenuItem>
      )}
      <MenuItem
        icon={<IconChecks />}
        onAction={() => {
          markAsRead(channelId);
        }}
      >
        {t("shell.channelMenu.markAsRead")}
      </MenuItem>
      <MenuItem
        icon={<IconLink />}
        onAction={() => {
          const { href } = router.buildLocation(location);
          void copyWithToast(toShareUrl(href), t("shell.channelMenu.linkCopied"));
        }}
      >
        {t("shell.channelMenu.copyLink")}
      </MenuItem>
      {!isTauri && (
        <MenuItemLink {...location} target="_blank" icon={<IconExternalLink />}>
          {t("shell.openInNewTab")}
        </MenuItemLink>
      )}
    </>
  );
};
