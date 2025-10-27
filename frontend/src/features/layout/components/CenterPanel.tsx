import { Text, ActionIcon } from "@mantine/core";
import { IconMenu2, IconInfoCircle } from "@tabler/icons-react";
import { Outlet, useParams } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import { useChannels } from "@/features/channel/hooks/useChannel";
import { MessageInput } from "@/features/message/components/MessageInput";
import {
  showLeftSidePanelAtom,
  showMobileLeftPanelAtom,
  showMobileRightPanelAtom,
} from "@/providers/store/ui";
import { setRightSidePanelViewAtom } from "@/providers/store/ui";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "@/providers/store/workspace";

export const CenterPanel = () => {
  const params = useParams({ strict: false });
  const channelId = params.channelId as string | undefined;
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const showLeftSidePanel = useSetAtom(showLeftSidePanelAtom);
  const showMobileLeftPanel = useSetAtom(showMobileLeftPanelAtom);
  const showMobileRightPanel = useSetAtom(showMobileRightPanelAtom);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  const { data: channels } = useChannels(currentWorkspaceId || "");
  const channel = channels?.find((c) => c.id === (channelId || currentChannelId));

  const handleLeftPanelToggle = () => {
    // デスクトップでは左パネルを表示、モバイルではモバイル左パネルを表示
    showLeftSidePanel();
    showMobileLeftPanel();
  };

  const handleRightPanelToggle = () => {
    // デスクトップでは右パネルを表示、モバイルではモバイル右パネルを表示
    setRightSidePanelView({ type: "channel-info", channelId: channelId || currentChannelId });
    showMobileRightPanel();
  };

  return (
    <div className="flex flex-col h-full min-h-0 bg-white">
      {/* CenterPanel ヘッダー */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 bg-white">
        <div className="flex items-center space-x-3">
          {/* モバイル用の左パネル切り替えボタン（CSSで表示制御） */}
          <ActionIcon
            variant="subtle"
            size="lg"
            onClick={handleLeftPanelToggle}
            className="text-gray-700 hover:bg-gray-100 md:hidden"
            title="チャンネル一覧"
          >
            <IconMenu2 size={20} />
          </ActionIcon>

          {/* チャンネル情報 */}
          <div className="flex-1 min-w-0">
            {channel ? (
              <div>
                <Text size="lg" fw={600} className="text-gray-900 truncate">
                  #{channel.name}
                </Text>
                {channel.description && (
                  <Text size="sm" c="dimmed" className="truncate">
                    {channel.description}
                  </Text>
                )}
              </div>
            ) : (
              <Text size="lg" fw={600} className="text-gray-900">
                チャンネルを選択
              </Text>
            )}
          </div>
        </div>

        <div className="flex items-center space-x-2">
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

      {/* メッセージ表示エリア */}
      <div className="flex-1 min-h-0">
        <Outlet />
      </div>

      {/* メッセージ入力エリア */}
      <MessageInput channelId={channelId || currentChannelId} />
    </div>
  );
};
