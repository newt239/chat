import { ChannelInfoPanel } from "./ChannelInfoPanel";
import { MemberPanel } from "./MemberPanel";
import { SearchResultsPanel } from "./SearchResultsPanel";
import { ThreadPanel } from "./ThreadPanel";
import { UserProfilePanel } from "./UserProfilePanel";

import { BookmarkList } from "@/features/bookmark/components/BookmarkList";
import { type PanelView } from "@/providers/store/ui";

type WorkspaceRightSidebarProps = {
  workspaceId: string | null;
  view: PanelView;
};

export const WorkspaceRightSidebar = ({ workspaceId, view }: WorkspaceRightSidebarProps) => {
  if (!workspaceId) {
    return null;
  }

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
    case "hidden":
      return null;
  }
};
