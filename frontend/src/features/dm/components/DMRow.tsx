import { ChannelNavItem } from "#/features/channel/components/ChannelNavItem";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";

import { dmName } from "../utils/dmName";
import { DMAvatar } from "./DMAvatar";

import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";

type DMRowProps = {
  workspaceId: string;
  dm: DirectMessage;
};

export const DMRow = ({ workspaceId, dm }: DMRowProps) => {
  const displayName = useDisplayName();
  return (
    <ChannelNavItem
      workspaceId={workspaceId}
      channelId={dm.id}
      isStarred={dm.isStarred}
      isMuted={dm.isMuted}
      unreadCount={dm.unreadCount}
      showsBadge
    >
      <DMAvatar dm={dm} size={18} />
      <span className="min-w-0 flex-1 truncate">{dmName(dm, displayName)}</span>
    </ChannelNavItem>
  );
};
