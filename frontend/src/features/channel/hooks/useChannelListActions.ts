import { timestampNow } from "@bufbuild/protobuf/wkt";
import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { channelListKey } from "#/features/channel/hooks/useChannel";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { ReadStateService } from "#/gen/chat/v1/read_state_service_pb";

/** サイドバーの行やチャンネルの「その他」メニューから行う、スター・ミュート・既読の操作 */
export const useChannelListActions = (workspaceId: string) => {
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) }),
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          schema: DirectMessageService.method.listDirectMessages,
        }),
      }),
    ]);
  };
  const setStarred = useMutation(ChannelService.method.setChannelStarred, { onSuccess });
  const setMuted = useMutation(ChannelService.method.setChannelMuted, { onSuccess });
  const updateReadState = useMutation(ReadStateService.method.updateReadState, { onSuccess });

  return {
    markAsRead: (channelId: string) => {
      updateReadState.mutate({ channelId, lastReadAt: timestampNow() });
    },
    setMuted: (channelId: string, muted: boolean) => {
      setMuted.mutate({ channelId, muted });
    },
    setStarred: (channelId: string, starred: boolean) => {
      setStarred.mutate({ channelId, starred });
    },
  };
};
