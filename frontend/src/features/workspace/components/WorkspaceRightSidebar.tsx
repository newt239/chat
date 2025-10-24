import { ChannelInfoPanel } from "./ChannelInfoPanel";
import { MemberPanel } from "./MemberPanel";
import { SearchResultsPanel } from "./SearchResultsPanel";
import { ThreadPanel } from "./ThreadPanel";
import { UserProfilePanel } from "./UserProfilePanel";

import { type RightSidebarView } from "@/lib/store/ui";

type WorkspaceRightSidebarProps = {
  workspaceId: string;
  view: RightSidebarView;
};

export const WorkspaceRightSidebar = ({ workspaceId, view }: WorkspaceRightSidebarProps) => {
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
    case "hidden":
      return null;
  }
};
