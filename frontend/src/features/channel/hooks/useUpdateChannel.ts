import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { channelListKey } from "#/features/channel/hooks/useChannel";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

export const useUpdateChannel = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation(ChannelService.method.updateChannel, {
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) });
    },
  });
};
