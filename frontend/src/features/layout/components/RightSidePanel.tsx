import { ActionIcon, ScrollArea } from "@mantine/core";
import { IconX } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import type { PanelView } from "@/providers/store/ui";

import { BookmarkList } from "@/features/bookmark/components/BookmarkList";
import { NotificationPanel } from "@/features/notification/components/NotificationPanel";
import { ChannelInfoPanel } from "@/features/sidebar/components/ChannelInfoPanel";
import { MemberPanel } from "@/features/sidebar/components/MemberPanel";
import { SearchResultsPanel } from "@/features/sidebar/components/SearchResultsPanel";
import { ThreadPanel } from "@/features/sidebar/components/ThreadPanel";
import { UserProfilePanel } from "@/features/sidebar/components/UserProfilePanel";
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
      case "members":
        return <MemberPanel workspaceId={workspaceId} />;
      case "channel-info":
        return <ChannelInfoPanel workspaceId={workspaceId} channelId={view.channelId} />;
      case "thread":
        return <ThreadPanel threadId={view.threadId} />;
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
      case "members":
        return "メンバー";
      case "channel-info":
        return "チャンネル情報";
      case "thread":
        return "スレッド";
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
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200">
        <h3 className="text-lg font-semibold text-gray-900">{getPanelTitle(rightSidePanelView)}</h3>
        <ActionIcon
          variant="subtle"
          size="sm"
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
