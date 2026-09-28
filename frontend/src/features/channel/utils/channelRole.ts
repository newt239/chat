import { ChannelRole } from "#/gen/chat/v1/channel_member_service_pb";

export const channelRoleLabels: Record<ChannelRole, string> = {
  [ChannelRole.UNSPECIFIED]: "",
  [ChannelRole.MEMBER]: "メンバー",
  [ChannelRole.ADMIN]: "管理者",
};
