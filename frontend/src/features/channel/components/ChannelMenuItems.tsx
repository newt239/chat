import {
  IconBell,
  IconBellOff,
  IconChecks,
  IconExternalLink,
  IconLink,
  IconStar,
  IconStarOff,
} from "@tabler/icons-react";
import { useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useChannelListActions } from "#/features/channel/hooks/useChannelListActions";
import { toShareUrl } from "#/lib/platform/appOrigin";

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
  const { markAsRead, setMuted, setStarred } = useChannelListActions(workspaceId);
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
          void navigator.clipboard.writeText(toShareUrl(href));
          toast(t("shell.channelMenu.linkCopied"));
        }}
      >
        {t("shell.channelMenu.copyLink")}
      </MenuItem>
      <MenuItemLink {...location} target="_blank" icon={<IconExternalLink />}>
        {t("shell.openInNewTab")}
      </MenuItemLink>
    </>
  );
};
