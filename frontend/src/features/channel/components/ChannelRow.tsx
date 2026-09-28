import { ChannelName } from "./ChannelName";
import { ChannelNavItem } from "./ChannelNavItem";

import type { Channel } from "#/gen/chat/v1/channel_service_pb";

type ChannelRowProps = {
  workspaceId: string;
  channel: Channel;
};

export const ChannelRow = ({ workspaceId, channel }: ChannelRowProps) => (
  <ChannelNavItem
    workspaceId={workspaceId}
    channelId={channel.id}
    isStarred={channel.isStarred}
    isMuted={channel.isMuted}
    unreadCount={channel.unreadCount}
    showsBadge={channel.hasMention}
  >
    <ChannelName name={channel.name} isPrivate={channel.isPrivate} />
  </ChannelNavItem>
);
