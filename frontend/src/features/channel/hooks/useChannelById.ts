import { skipToken, useQuery } from "@connectrpc/connect-query";

import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { useChannels } from "./useChannel";

// 一覧になければ個別に取得する。channelId が null なら取得しない
export const useChannelById = (workspaceId: string, channelId: string | null) => {
  const list = useChannels(workspaceId);
  const listed = list.data?.find((channel) => channel.id === channelId);
  const fetched = useQuery(
    ChannelService.method.getChannel,
    list.data !== undefined && listed === undefined && channelId !== null
      ? { channelId }
      : skipToken,
    { select: (res) => res.channel },
  );
  return {
    channel: listed ?? fetched.data,
    isError: list.isError || fetched.isError,
    isPending: listed === undefined && fetched.isPending,
  };
};
