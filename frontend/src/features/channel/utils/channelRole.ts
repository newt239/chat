import { ChannelRole } from "#/gen/chat/v1/channel_member_service_pb";

export const channelRoleKeys = {
  [ChannelRole.UNSPECIFIED]: "channel.roles.member",
  [ChannelRole.MEMBER]: "channel.roles.member",
  [ChannelRole.ADMIN]: "channel.roles.admin",
} as const;
