import { skipToken, useQuery } from "@connectrpc/connect-query";
import {
  IconBellOff,
  IconNote,
  IconDots,
  IconInfoCircle,
  IconPin,
  IconStar,
  IconStarFilled,
  IconUser,
  IconUsers,
} from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { BackButton } from "#/components/block/BackButton/BackButton";
import { useMobileForward } from "#/components/block/BackButton/mobileStack";
import { Avatar } from "#/components/ui/Avatar/Avatar";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { IconToggleButton } from "#/components/ui/IconToggleButton/IconToggleButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { focusRing } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";
import { DMAvatar } from "#/features/channel/components/DMAvatar";
import { dmName } from "#/features/channel/utils/dmName";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { usePinCount } from "#/features/pin/hooks/usePinnedMessages";
import { DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { useIsMobile } from "#/hooks/useMediaQuery";
import { openPanel } from "#/lib/overlaySearch";

import { useChannelAggregation } from "../hooks/useChannelAggregation";
import { useChannelListActions } from "../hooks/useChannelListActions";
import { useChannelMembers } from "../hooks/useChannelMembers";
import { ChannelLinkBar } from "./ChannelLinkBar";
import { ChannelMenuItems } from "./ChannelMenuItems";
import { ChannelName } from "./ChannelName";
import { DescendantsToggle } from "./DescendantsToggle";

import type { Channel } from "#/gen/chat/v1/channel_service_pb";
import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";
import type { PanelSearch } from "#/lib/overlaySearch";

type ChannelHeaderProps = {
  workspaceId: string;
  channelId: string;
  // DM のときは channel が、チャンネルのときは dm が undefined
  channel: Channel | undefined;
  dm: DirectMessage | undefined;
};

export const ChannelHeader = ({ workspaceId, channelId, channel, dm }: ChannelHeaderProps) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const pinsCount = usePinCount(channelId);
  const { descendants, includesDescendants, setIncludesDescendants } = useChannelAggregation(
    workspaceId,
    channelId,
  );
  const { data: members = [] } = useChannelMembers(channelId);
  const { setStarred } = useChannelListActions(workspaceId);
  const displayName = useDisplayName();

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
  // モバイルでは左へのスワイプでも同じ情報を開く
  useMobileForward(() => {
    openRightPanel(infoPanel);
  });

  if (!channel && !dm) {
    return <header className="h-12 shrink-0 border-b border-border" />;
  }

  return (
    <>
      <header className="flex h-12 shrink-0 items-center gap-1.5 border-b border-border pr-2.5 pl-4.5 max-md:pl-3">
        <BackButton />
        <IconToggleButton
          label={t("shell.channelMenu.star")}
          isSelected={isStarred}
          className="max-md:hidden data-selected:bg-transparent data-selected:text-mention-bar data-selected:data-hovered:bg-hover"
          onChange={(starred) => {
            setStarred(channelId, starred);
          }}
        >
          {isStarred ? <IconStarFilled /> : <IconStar />}
        </IconToggleButton>
        <Button
          onPress={() => {
            openRightPanel(infoPanel);
          }}
          className={`flex min-w-0 shrink cursor-pointer items-center gap-1 rounded-md px-1 py-0.5 text-title font-bold whitespace-nowrap data-hovered:bg-hover [&_svg]:size-4 [&_svg]:shrink-0 [&_svg]:text-muted ${focusRing}`}
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
        <p className="m-0 flex min-w-0 flex-1 items-center gap-1 truncate pl-1.5 text-label font-normal text-muted max-md:invisible [&_svg]:size-3.5 [&_svg]:shrink-0">
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
            className={`flex h-7 shrink-0 cursor-pointer items-center gap-1.5 rounded-md border max-md:hidden border-border py-0.5 pr-2 pl-0.75 text-xs text-muted tabular-nums data-hovered:bg-hover ${focusRing}`}
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
            <span className="absolute top-px right-0 font-mono text-caption leading-none font-semibold text-muted">
              {pinsCount > 99 ? "99+" : pinsCount}
            </span>
          )}
        </IconButton>
        {channel && (
          <IconButton
            label={t("shell.rightPanel.channelInfo")}
            onPress={() => {
              openRightPanel({ panel: "info" });
            }}
          >
            <IconInfoCircle />
          </IconButton>
        )}
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
