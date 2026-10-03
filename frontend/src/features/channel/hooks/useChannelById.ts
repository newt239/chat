import { skipToken, useQuery } from "@connectrpc/connect-query";

import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { useChannels } from "./useChannel";

// 一覧になければ個別に取得する。channelId が null なら取得しない
export const useChannelById = (workspaceId: string, channelId: string | null) => {
  const { data: channels } = useChannels(workspaceId);
  const listed = channels?.find((channel) => channel.id === channelId);
  const { data: fetched } = useQuery(
    ChannelService.method.getChannel,
    channels !== undefined && listed === undefined && channelId !== null
      ? { channelId }
      : skipToken,
    { select: (res) => res.channel },
  );
  return listed ?? fetched;
};
