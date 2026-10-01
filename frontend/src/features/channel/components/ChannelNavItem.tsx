import type { ReactNode } from "react";

import { IconBellOff } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { NavLink } from "#/components/block/NavLink/NavLink";
import { Badge } from "#/components/ui/Badge/Badge";
import { ContextMenu } from "#/components/ui/ContextMenu/ContextMenu";
import { cn } from "#/components/ui/styles/styles";
import { DraftIndicator } from "#/features/draft/components/DraftIndicator";

import { ChannelMenuItems } from "./ChannelMenuItems";

type ChannelNavItemProps = {
  workspaceId: string;
  channelId: string;
  isStarred: boolean;
  isMuted: boolean;
  unreadCount: number;
  // バッジに出す件数（チャンネルはメンション数、DM は未読数）。0 なら出さない
  badgeCount: number;
  // アイコンと名前
  children: ReactNode;
};

// サイドバーのチャンネル・DM の行。右クリックでスターや既読などの操作を出す
export const ChannelNavItem = ({
  workspaceId,
  channelId,
  isStarred,
  isMuted,
  unreadCount,
  badgeCount,
  children,
}: ChannelNavItemProps) => {
  const { t } = useTranslation();
  // ミュート中は未読を強調しない
  const hasUnread = unreadCount > 0 && !isMuted;
  return (
    <ContextMenu
      aria-label={t("shell.channelMenu.label")}
      menu={
        <ChannelMenuItems
          workspaceId={workspaceId}
          channelId={channelId}
          isStarred={isStarred}
          isMuted={isMuted}
        />
      }
    >
      <NavLink
        to="/app/$workspaceId/$channelId"
        params={{ channelId, workspaceId }}
        className={cn(
          hasUnread && "font-semibold text-(--nav-strong) [&_svg]:text-(--nav-strong)",
          isMuted && "opacity-55",
        )}
      >
        {children}
        <DraftIndicator workspaceId={workspaceId} channelId={channelId} />
        {isMuted && <IconBellOff aria-label={t("shell.channel.muted")} role="img" />}
        {badgeCount > 0 && <Badge>{badgeCount > 99 ? "99+" : badgeCount}</Badge>}
      </NavLink>
    </ContextMenu>
  );
};
