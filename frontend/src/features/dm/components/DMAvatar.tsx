import { Avatar } from "#/components/ui/Avatar";
import { GroupAvatar } from "#/components/ui/GroupAvatar";
import { DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";

import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";

type DMAvatarProps = {
  dm: DirectMessage;
  size: number;
};

export const DMAvatar = ({ dm, size }: DMAvatarProps) => {
  const [member] = dm.members;
  if (dm.type === DirectMessageType.GROUP_DM || member === undefined) {
    return <GroupAvatar count={dm.members.length + 1} size={size} />;
  }
  return <Avatar name={member.displayName} src={member.avatarUrl} size={size} />;
};
