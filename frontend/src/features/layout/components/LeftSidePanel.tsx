import { useState } from "react";

import { Text, ActionIcon, ScrollArea } from "@mantine/core";
import { IconX, IconPlus } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import { ChannelList } from "@/features/channel/components/ChannelList";
import { CreateChannelModal } from "@/features/channel/components/CreateChannelModal";
import { CreateDMModal } from "@/features/dm/components/CreateDMModal";
import { DMList } from "@/features/dm/components/DMList";
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
  const [isCreateDMModalOpen, setIsCreateDMModalOpen] = useState(false);

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

  const handleCreateDM = () => {
    setIsCreateDMModalOpen(true);
  };

  return (
    <div className={`bg-white border-r border-gray-200 flex flex-col h-full ${className}`}>
      {/* チャンネルとDM一覧 */}
      <div className="flex-1 min-h-0">
        <ScrollArea className="h-full">
          <div className="px-2 py-3 space-y-4">
            {/* チャンネルセクション */}
            <div>
              <div className="flex items-center justify-between px-2 mb-2">
                <Text size="sm" fw={600} c="dimmed">
                  チャンネル
                </Text>
                <ActionIcon
                  variant="subtle"
                  size="sm"
                  onClick={handleCreateChannel}
                  className="text-gray-500 hover:bg-gray-100"
                >
                  <IconPlus size={14} title="チャンネルを作成" />
                </ActionIcon>
                <div className="md:hidden">
                  <ActionIcon
                    variant="subtle"
                    size="lg"
                    onClick={handleClose}
                    className="text-gray-500 hover:bg-gray-100"
                  >
                    <IconX size={16} title="閉じる" />
                  </ActionIcon>
                </div>
              </div>
              <ChannelList workspaceId={currentWorkspaceId} />
            </div>

            {/* DMセクション */}
            <div>
              <div className="flex items-center justify-between px-2 mb-2">
                <Text size="sm" fw={600} c="dimmed">
                  ダイレクトメッセージ
                </Text>
                <ActionIcon
                  variant="subtle"
                  size="sm"
                  onClick={handleCreateDM}
                  className="text-gray-500 hover:bg-gray-100"
                  title="ダイレクトメッセージを作成"
                >
                  <IconPlus size={14} />
                </ActionIcon>
              </div>
              {currentWorkspaceId && <DMList workspaceId={currentWorkspaceId} />}
            </div>
          </div>
        </ScrollArea>
      </div>

      {/* チャンネル作成モーダル */}
      <CreateChannelModal
        workspaceId={currentWorkspaceId}
        opened={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />

      {/* DM作成モーダル */}
      {currentWorkspaceId && (
        <CreateDMModal
          workspaceId={currentWorkspaceId}
          opened={isCreateDMModalOpen}
          onClose={() => setIsCreateDMModalOpen(false)}
        />
      )}
    </div>
  );
};
