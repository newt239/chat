import { useQuery } from "@connectrpc/connect-query";

import { ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";

export const useChannelMembers = (channelId: string) =>
  useQuery(
    ChannelMemberService.method.listChannelMembers,
    { channelId },
    { select: (res) => res.members },
  );
