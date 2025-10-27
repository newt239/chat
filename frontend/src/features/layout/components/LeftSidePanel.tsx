import { Text, ActionIcon, ScrollArea } from "@mantine/core";
import { IconX, IconPlus } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import { ChannelList } from "@/features/channel/components/ChannelList";
import { leftSidePanelVisibleAtom, hideMobilePanelsAtom } from "@/providers/store/ui";
import { currentWorkspaceIdAtom } from "@/providers/store/workspace";

type LeftSidePanelProps = {
  className?: string;
};

export const LeftSidePanel = ({ className = "" }: LeftSidePanelProps) => {
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const leftSidePanelVisible = useAtomValue(leftSidePanelVisibleAtom);
  const hideMobilePanels = useSetAtom(hideMobilePanelsAtom);

  // デスクトップで左パネルが非表示の場合は非表示
  if (!leftSidePanelVisible) {
    return null;
  }

  const handleClose = () => {
    // モバイルでのみ閉じるボタンが表示されるため、常にモバイルパネルを閉じる
    hideMobilePanels();
  };

  const handleCreateChannel = () => {
    // TODO: チャンネル作成モーダルを開く
    console.log("Create channel clicked");
  };

  return (
    <div className={`bg-white border-r border-gray-200 flex flex-col h-full ${className}`}>
      {/* ヘッダー */}
      <div className="flex items-center justify-between px-4 py-3">
        <Text size="lg" fw={600} className="text-gray-900">
          チャンネル
        </Text>
        <div className="flex items-center space-x-2">
          <ActionIcon
            variant="subtle"
            size="sm"
            onClick={handleCreateChannel}
            className="text-gray-500 hover:bg-gray-100"
          >
            <IconPlus size={16} />
          </ActionIcon>
          {/* モバイル用の閉じるボタン（CSSで表示制御） */}
          <ActionIcon
            variant="subtle"
            size="sm"
            onClick={handleClose}
            className="text-gray-500 hover:bg-gray-100 md:hidden"
          >
            <IconX size={16} />
          </ActionIcon>
        </div>
      </div>

      {/* チャンネル一覧 */}
      <div className="flex-1 min-h-0">
        <ScrollArea className="h-full">
          <ChannelList workspaceId={currentWorkspaceId} />
        </ScrollArea>
      </div>
    </div>
  );
};
