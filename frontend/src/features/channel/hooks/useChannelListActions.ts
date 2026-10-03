import { timestampNow } from "@bufbuild/protobuf/wkt";
import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { channelListKey } from "./useChannel";
import { updateListedDM } from "./useDM";
import { useUpdateReadState } from "./useUpdateReadState";

import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";

/** サイドバーの行やチャンネルの「その他」メニューから行う、スター・ミュート・既読の操作 */
export const useChannelListActions = (workspaceId: string) => {
  const queryClient = useQueryClient();
  const onSuccess = async (channelId: string, update: (dm: DirectMessage) => DirectMessage) => {
    updateListedDM(queryClient, { channelId, workspaceId }, update);
    await queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) });
  };
  const setStarred = useMutation(ChannelService.method.setChannelStarred, {
    onSuccess: (_, { channelId = "", starred = false }) =>
      onSuccess(channelId, (dm) => ({ ...dm, isStarred: starred })),
  });
  const setMuted = useMutation(ChannelService.method.setChannelMuted, {
    onSuccess: (_, { channelId = "", muted = false }) =>
      onSuccess(channelId, (dm) => ({ ...dm, isMuted: muted })),
  });
  const updateReadState = useUpdateReadState(workspaceId);

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
