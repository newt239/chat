import { ActionIcon, Badge, Group, Tooltip } from "@mantine/core";
import { IconInfoCircle, IconMenu2, IconUsers, IconPin } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import { ChannelName } from "./ChannelName";

import { useChannels } from "@/features/channel/hooks/useChannel";
import {
  showLeftSidePanelAtom,
  showMobileLeftPanelAtom,
  showMobileRightPanelAtom,
  pinsCountByChannelAtom,
  setRightSidePanelViewAtom,
} from "@/providers/store/ui";
import { currentWorkspaceIdAtom } from "@/providers/store/workspace";

type ChannelHeaderProps = {
  channelId: string | null;
};

export const ChannelHeader = ({ channelId }: ChannelHeaderProps) => {
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);

  const showLeftSidePanel = useSetAtom(showLeftSidePanelAtom);
  const showMobileLeftPanel = useSetAtom(showMobileLeftPanelAtom);
  const showMobileRightPanel = useSetAtom(showMobileRightPanelAtom);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);
  const pinsCountByChannel = useAtomValue(pinsCountByChannelAtom);

  const { data: channels } = useChannels(currentWorkspaceId || "");

  if (!channelId) {
    return null;
  }

  const channel = channels?.find((c) => c.id === channelId);

  const handleLeftPanelToggle = () => {
    // デスクトップでは左パネルを表示、モバイルではモバイル左パネルを表示
    showLeftSidePanel();
    showMobileLeftPanel();
  };

  const handleMembersPanelToggle = () => {
    setRightSidePanelView({ type: "channel-members", channelId });
    showMobileRightPanel();
  };

  const handleRightPanelToggle = () => {
    // デスクトップでは右パネルを表示、モバイルではモバイル右パネルを表示
    setRightSidePanelView({ type: "channel-info", channelId: channelId });
    showMobileRightPanel();
  };

  const handlePinsPanelOpen = () => {
    if (!channelId) return;
    setRightSidePanelView({ type: "pins", channelId });
    showMobileRightPanel();
  };

  return (
    <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 bg-white">
      <div className="flex items-center space-x-3">
        {/* モバイル用の左パネル切り替えボタン（CSSで表示制御） */}
        <div className="md:hidden">
          <ActionIcon
            variant="subtle"
            size="lg"
            onClick={handleLeftPanelToggle}
            className="text-gray-700 hover:bg-gray-100 md:hidden"
            title="チャンネル一覧"
          >
            <IconMenu2 size={20} />
          </ActionIcon>
        </div>

        {/* チャンネル情報 */}
        <div className="flex-1 min-w-0">
          {channel && (
            <div>
              <ChannelName name={channel.name} isPrivate={channel.isPrivate} />
              {channel.description && (
                <p className="text-sm text-gray-500 truncate">{channel.description}</p>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center space-x-2">
        <Tooltip label="ピン留め一覧" withArrow>
          <ActionIcon
            variant="subtle"
            size="lg"
            onClick={handlePinsPanelOpen}
            className="text-gray-700 hover:bg-gray-100 relative"
            title="ピン留め一覧"
          >
            <Group gap={4} align="center">
              <IconPin size={20} />
              {channelId && (pinsCountByChannel[channelId] ?? 0) > 0 && (
                <Badge size="xs" color="grape" variant="filled">
                  {(pinsCountByChannel[channelId] ?? 0) > 99
                    ? "99+"
                    : pinsCountByChannel[channelId]}
                </Badge>
              )}
            </Group>
          </ActionIcon>
        </Tooltip>
        <ActionIcon
          variant="subtle"
          size="lg"
          onClick={handleMembersPanelToggle}
          className="text-gray-700 hover:bg-gray-100"
          title="メンバー"
        >
          <IconUsers size={20} />
        </ActionIcon>
        {/* チャンネル情報ボタン */}
        <ActionIcon
          variant="subtle"
          size="lg"
          onClick={handleRightPanelToggle}
          className="text-gray-700 hover:bg-gray-100"
          title="チャンネル情報"
        >
          <IconInfoCircle size={20} />
        </ActionIcon>
      </div>
    </div>
  );
};
