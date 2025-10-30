import { ActionIcon, ScrollArea } from "@mantine/core";
import { IconX } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import type { PanelView } from "@/providers/store/ui";

import { BookmarkList } from "@/features/bookmark/components/BookmarkList";
import { ChannelInfoPanel } from "@/features/channel/components/ChannelInfoPanel";
import { ChannelMemberPanel } from "@/features/channel/components/ChannelMemberPanel";
import { UserProfilePanel } from "@/features/member/components/UserProfilePanel";
import { NotificationPanel } from "@/features/notification/components/NotificationPanel";
import { PinnedPanel } from "@/features/pin/components/PinnedPanel";
import { SearchResultsPanel } from "@/features/search/components/SearchResultsPanel";
import { ThreadPanel } from "@/features/thread/components/ThreadPanel";
import {
  rightSidePanelViewAtom,
  hideMobilePanelsAtom,
  closeRightSidePanelAtom,
} from "@/providers/store/ui";
import { currentWorkspaceIdAtom } from "@/providers/store/workspace";

type RightSidePanelProps = {
  className?: string;
};

export const RightSidePanel = ({ className = "" }: RightSidePanelProps) => {
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
  const rightSidePanelView = useAtomValue(rightSidePanelViewAtom);
  const hideMobilePanels = useSetAtom(hideMobilePanelsAtom);
  const closeRightSidePanel = useSetAtom(closeRightSidePanelAtom);

  // 右パネルが非表示の場合は何も表示しない
  if (rightSidePanelView.type === "hidden") {
    return null;
  }

  if (!workspaceId) {
    return null;
  }

  const handleClose = () => {
    // デスクトップでは右パネルを閉じる、モバイルではモバイルパネルを閉じる
    closeRightSidePanel();
    hideMobilePanels();
  };

  const renderPanelContent = (view: PanelView) => {
    switch (view.type) {
      case "channel-members":
        return <ChannelMemberPanel channelId={view.channelId} />;
      case "channel-info":
        return <ChannelInfoPanel workspaceId={workspaceId} channelId={view.channelId} />;
      case "thread":
        return <ThreadPanel threadId={view.threadId} />;
      case "pins":
        return <PinnedPanel channelId={view.channelId} />;
      case "user-profile":
        return <UserProfilePanel workspaceId={workspaceId} userId={view.userId} />;
      case "search":
        return (
          <SearchResultsPanel workspaceId={workspaceId} query={view.query} filter={view.filter} />
        );
      case "bookmarks":
        return <BookmarkList />;
      case "notifications":
        return <NotificationPanel />;
      case "hidden":
        return null;
    }
  };

  const getPanelTitle = (view: PanelView) => {
    switch (view.type) {
      case "channel-members":
        return "メンバー";
      case "channel-info":
        return "チャンネル情報";
      case "thread":
        return "スレッド";
      case "pins":
        return "ピン留め";
      case "user-profile":
        return "ユーザープロフィール";
      case "search":
        return "検索結果";
      case "bookmarks":
        return "ブックマーク";
      case "notifications":
        return "通知";
      case "hidden":
        return "";
    }
  };

  return (
    <div className={`bg-white border-l border-gray-200 flex flex-col h-full ${className}`}>
      {/* ヘッダー */}
      <div className="flex items-center justify-between px-4 py-3">
        <h3 className="text-lg font-semibold text-gray-900">{getPanelTitle(rightSidePanelView)}</h3>
        <ActionIcon
          variant="subtle"
          size="lg"
          onClick={handleClose}
          className="text-gray-500 hover:bg-gray-100"
        >
          <IconX size={16} />
        </ActionIcon>
      </div>

      {/* パネル内容 */}
      <div className="flex-1 min-h-0">
        <ScrollArea className="h-full">{renderPanelContent(rightSidePanelView)}</ScrollArea>
      </div>
    </div>
  );
};
