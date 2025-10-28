import { useState } from "react";

import { Text, ActionIcon, ScrollArea } from "@mantine/core";
import { IconX, IconPlus } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import { ChannelList } from "@/features/channel/components/ChannelList";
import { CreateChannelModal } from "@/features/channel/components/CreateChannelModal";
import { leftSidePanelVisibleAtom, hideMobilePanelsAtom } from "@/providers/store/ui";
import { currentWorkspaceIdAtom } from "@/providers/store/workspace";

type LeftSidePanelProps = {
  className?: string;
};

export const LeftSidePanel = ({ className = "" }: LeftSidePanelProps) => {
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const leftSidePanelVisible = useAtomValue(leftSidePanelVisibleAtom);
  const hideMobilePanels = useSetAtom(hideMobilePanelsAtom);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);

  // デスクトップで左パネルが非表示の場合は非表示
  if (!leftSidePanelVisible) {
    return null;
  }

  const handleClose = () => {
    // モバイルでのみ閉じるボタンが表示されるため、常にモバイルパネルを閉じる
    hideMobilePanels();
  };

  const handleCreateChannel = () => {
    setIsCreateModalOpen(true);
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
            size="lg"
            onClick={handleCreateChannel}
            className="text-gray-500 hover:bg-gray-100"
          >
            <IconPlus size={16} title="チャンネルを作成" />
          </ActionIcon>
          <div className="md:hidden">
            <ActionIcon
              variant="subtle"
              size="lg"
              onClick={handleClose}
              className="text-gray-500 hover:bg-gray-100 md:hidden"
            >
              <IconX size={16} title="チャンネル一覧を閉じる" />
            </ActionIcon>
          </div>
        </div>
      </div>

      {/* チャンネル一覧 */}
      <div className="flex-1 min-h-0">
        <ScrollArea className="h-full">
          <ChannelList workspaceId={currentWorkspaceId} />
        </ScrollArea>
      </div>

      {/* チャンネル作成モーダル */}
      <CreateChannelModal
        workspaceId={currentWorkspaceId}
        opened={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />
    </div>
  );
};
