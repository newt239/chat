import { createConnectQueryKey, skipToken, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

export const channelListKey = (workspaceId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { workspaceId },
    schema: ChannelService.method.listChannels,
  });

export const useChannels = (workspaceId: string | null) =>
  useQuery(ChannelService.method.listChannels, workspaceId === null ? skipToken : { workspaceId }, {
    select: (res) => res.channels,
  });

export const useCreateChannel = () => {
  const queryClient = useQueryClient();

  return useMutation(ChannelService.method.createChannel, {
    onSuccess: async (_, { workspaceId = "" }) => {
      await queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) });
    },
  });
};
