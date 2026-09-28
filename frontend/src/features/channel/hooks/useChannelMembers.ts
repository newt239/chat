import { skipToken, useQuery } from "@connectrpc/connect-query";

import { ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";

export const useChannelMembers = (channelId: string | null) =>
  useQuery(
    ChannelMemberService.method.listChannelMembers,
    channelId === null ? skipToken : { channelId },
    { select: (res) => res.members },
  );
