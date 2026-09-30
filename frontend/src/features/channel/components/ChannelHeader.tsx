import { skipToken, useQuery } from "@connectrpc/connect-query";
import {
  IconBellOff,
  IconNote,
  IconChevronDown,
  IconDots,
  IconInfoCircle,
  IconPin,
  IconStar,
  IconStarFilled,
  IconUser,
  IconUsers,
} from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { focusRing } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";
import { DMAvatar } from "#/features/dm/components/DMAvatar";
import { useDMs } from "#/features/dm/hooks/useDM";
import { dmName } from "#/features/dm/utils/dmName";
import { BackButton } from "#/features/layout/components/BackButton";
import { openPanel } from "#/features/layout/utils/overlaySearch";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { useIsMobile } from "#/hooks/useMediaQuery";
import { pinsCountByChannelAtom } from "#/providers/store/ui";

import { useChannelAggregation } from "../hooks/useChannelAggregation";
import { useChannelListActions } from "../hooks/useChannelListActions";
import { useChannelMembers } from "../hooks/useChannelMembers";
import { ChannelLinkBar } from "./ChannelLinkBar";
import { ChannelMenuItems } from "./ChannelMenuItems";
import { ChannelName } from "./ChannelName";
import { DescendantsToggle } from "./DescendantsToggle";

import type { PanelSearch } from "#/features/layout/utils/overlaySearch";

type ChannelHeaderProps = {
  workspaceId: string;
  channelId: string;
};

export const ChannelHeader = ({ workspaceId, channelId }: ChannelHeaderProps) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const pinsCount = useAtomValue(pinsCountByChannelAtom)[channelId] ?? 0;
  const { channel, descendants, includesDescendants, setIncludesDescendants } =
    useChannelAggregation(workspaceId, channelId);
  const { data: dms } = useDMs(workspaceId);
  const { data: members = [] } = useChannelMembers(channelId);
  const { setStarred } = useChannelListActions(workspaceId);
  const displayName = useDisplayName();

  const dm = dms?.find((candidate) => candidate.id === channelId);
  const isStarred = channel?.isStarred ?? dm?.isStarred ?? false;
  const isMuted = channel?.isMuted ?? dm?.isMuted ?? false;
  const isGroupDM = dm?.type === DirectMessageType.GROUP_DM;
  const [partner] = dm?.type === DirectMessageType.DM ? dm.members : [];
  // 1 対 1 の DM では相手に付けたメモをトピックの位置に出す
  const { data: memo } = useQuery(
    UserService.method.getUserNote,
    partner ? { targetUserId: partner.userId } : skipToken,
    { select: (res) => res.note?.memo ?? "" },
  );
  const descendantsToggle = descendants.length > 0 && (
    <DescendantsToggle
      count={descendants.length}
      isSelected={includesDescendants}
      onChange={setIncludesDescendants}
    />
  );
  const openRightPanel = (panel: PanelSearch) => {
    void navigate({ search: openPanel(panel), to: "." });
  };
  // タイトルを押すと、1 対 1 の DM は相手のプロフィール、グループ DM はメンバー、チャンネルは情報を開く
  const infoPanel: PanelSearch = partner
    ? { profile: partner.userId }
    : { panel: dm ? "members" : "info" };

  if (!channel && !dm) {
    return <header className="h-12 shrink-0 border-b border-border" />;
  }

  return (
    <>
      <header className="flex h-12 shrink-0 items-center gap-1.5 border-b border-border pr-2.5 pl-[18px] max-md:pl-3">
        <BackButton />
        <IconButton
          label={isStarred ? t("shell.channelMenu.unstar") : t("shell.channelMenu.star")}
          aria-pressed={isStarred}
          className={isStarred ? "text-mention-bar data-hovered:text-mention-bar" : undefined}
          onPress={() => {
            setStarred(channelId, !isStarred);
          }}
        >
          {isStarred ? <IconStarFilled /> : <IconStar />}
        </IconButton>
        <Button
          onPress={() => {
            openRightPanel(infoPanel);
          }}
          className={`flex min-w-0 shrink cursor-pointer items-center gap-1 rounded-[6px] px-1 py-0.5 text-[15px] font-bold whitespace-nowrap data-hovered:bg-hover [&_svg]:size-4 [&_svg]:shrink-0 [&_svg]:text-muted ${focusRing}`}
        >
          {channel ? (
            <ChannelName name={channel.name} isPrivate={channel.isPrivate} />
          ) : (
            dm && (
              <>
                <DMAvatar dm={dm} size={22} />
                <span className="min-w-0 truncate">{dmName(dm, displayName)}</span>
              </>
            )
          )}
          <IconChevronDown aria-hidden className="size-3!" />
        </Button>
        {isMuted && (
          <Tooltip content={t("shell.channel.muted")}>
            <Button
              aria-label={t("shell.channel.muted")}
              className="grid cursor-default place-items-center text-subtle [&_svg]:size-4"
            >
              <IconBellOff />
            </Button>
          </Tooltip>
        )}
        {!isMobile && descendantsToggle}
        <p className="m-0 flex min-w-0 flex-1 items-center gap-1 truncate pl-1.5 text-[12.5px] text-muted max-md:invisible [&_svg]:size-3.5 [&_svg]:shrink-0">
          {channel?.description}
          {isGroupDM && t("dm.header.groupCount", { count: dm.members.length + 1 })}
          {memo && (
            <>
              <IconNote aria-label={t("member.note.memo")} role="img" />
              <span className="truncate" title={memo}>
                {memo}
              </span>
            </>
          )}
        </p>
        {(channel !== undefined || isGroupDM) && members.length > 0 && (
          <Button
            aria-label={t("shell.rightPanel.members")}
            onPress={() => {
              openRightPanel({ panel: "members" });
            }}
            className={`flex h-7 shrink-0 cursor-pointer items-center gap-1.5 rounded-md border max-md:hidden border-border py-0.5 pr-2 pl-[3px] text-xs text-muted tabular-nums data-hovered:bg-hover ${focusRing}`}
          >
            <span className="flex [&>*+*]:-ml-1.5 [&>*]:ring-2 [&>*]:ring-surface">
              {members.slice(0, 3).map((member) => (
                <Avatar
                  key={member.userId}
                  name={displayName(member.userId, member.displayName)}
                  src={member.avatarUrl}
                  size={20}
                />
              ))}
            </span>
            {members.length}
          </Button>
        )}
        <IconButton
          label={t("shell.rightPanel.pins")}
          onPress={() => {
            openRightPanel({ panel: "pins" });
          }}
        >
          <IconPin />
          {pinsCount > 0 && (
            <span className="absolute top-px right-0 font-mono text-[9.5px] leading-none font-semibold text-muted">
              {pinsCount > 99 ? "99+" : pinsCount}
            </span>
          )}
        </IconButton>
        {partner && (
          <IconButton
            label={t("shell.rightPanel.profile")}
            onPress={() => {
              openRightPanel({ profile: partner.userId });
            }}
          >
            <IconUser />
          </IconButton>
        )}
        <Menu
          trigger={
            <IconButton label={t("shell.channelMenu.more")}>
              <IconDots />
            </IconButton>
          }
        >
          <ChannelMenuItems
            workspaceId={workspaceId}
            channelId={channelId}
            isStarred={isStarred}
            isMuted={isMuted}
          />
          <MenuSeparator />
          {channel && (
            <MenuItemLink icon={<IconInfoCircle />} to="." search={openPanel({ panel: "info" })}>
              {t("shell.rightPanel.channelInfo")}
            </MenuItemLink>
          )}
          <MenuItemLink icon={<IconUsers />} to="." search={openPanel({ panel: "members" })}>
            {t("shell.rightPanel.members")}
          </MenuItemLink>
        </Menu>
      </header>
      {isMobile && descendantsToggle && (
        <div className="flex shrink-0 items-center gap-2 border-b border-border px-3.5 py-1.5 text-xs text-muted">
          {descendantsToggle}
          {t("channel.aggregate.count", { count: descendants.length })}
        </div>
      )}
      {channel && !isMobile && <ChannelLinkBar channelId={channelId} />}
    </>
  );
};
