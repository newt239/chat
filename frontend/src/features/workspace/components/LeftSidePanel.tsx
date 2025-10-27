import { Button, Text, ActionIcon, Group, ScrollArea } from "@mantine/core";
import { IconX, IconPlus } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import { ChannelList } from "@/features/channel/components/ChannelList";
import {
  leftSidePanelVisibleAtom,
  isMobileAtom,
  mobileActivePanelAtom,
  hideMobilePanelsAtom,
} from "@/providers/store/ui";
import { currentWorkspaceIdAtom } from "@/providers/store/workspace";

type LeftSidePanelProps = {
  className?: string;
};

export const LeftSidePanel = ({ className = "" }: LeftSidePanelProps) => {
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const leftSidePanelVisible = useAtomValue(leftSidePanelVisibleAtom);
  const isMobile = useAtomValue(isMobileAtom);
  const mobileActivePanel = useAtomValue(mobileActivePanelAtom);
  const hideMobilePanels = useSetAtom(hideMobilePanelsAtom);

  // モバイルで左パネルがアクティブでない場合は非表示
  if (isMobile && mobileActivePanel !== "left") {
    return null;
  }

  // デスクトップで左パネルが非表示の場合は非表示
  if (!isMobile && !leftSidePanelVisible) {
    return null;
  }

  const handleClose = () => {
    if (isMobile) {
      hideMobilePanels();
    }
  };

  const handleCreateChannel = () => {
    // TODO: チャンネル作成モーダルを開く
    console.log("Create channel clicked");
  };

  const handleJoinChannel = () => {
    // TODO: チャンネル参加モーダルを開く
    console.log("Join channel clicked");
  };

  return (
    <div className={`bg-white border-r border-gray-200 flex flex-col h-full ${className}`}>
      {/* ヘッダー */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200">
        <Text size="lg" fw={600} className="text-gray-900">
          チャンネル
        </Text>
        <div className="flex items-center space-x-2">
          {/* モバイル用の閉じるボタン */}
          {isMobile && (
            <ActionIcon
              variant="subtle"
              size="sm"
              onClick={handleClose}
              className="text-gray-500 hover:bg-gray-100"
            >
              <IconX size={16} />
            </ActionIcon>
          )}
        </div>
      </div>

      {/* チャンネル作成・参加ボタン */}
      <div className="px-4 py-3 border-b border-gray-100">
        <Group gap="xs">
          <Button
            size="xs"
            variant="light"
            leftSection={<IconPlus size={14} />}
            onClick={handleCreateChannel}
            className="flex-1"
          >
            作成
          </Button>
          <Button size="xs" variant="outline" onClick={handleJoinChannel} className="flex-1">
            参加
          </Button>
        </Group>
      </div>

      {/* チャンネル一覧 */}
      <div className="flex-1 min-h-0">
        <ScrollArea className="h-full">
          <div className="p-2">
            <ChannelList workspaceId={currentWorkspaceId} />
          </div>
        </ScrollArea>
      </div>
    </div>
  );
};
