import { timestampNow } from "@bufbuild/protobuf/wkt";
import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { channelListKey } from "./useChannel";
import { dmListKey } from "./useDM";
import { useUpdateReadState } from "./useUpdateReadState";

import type {
  DirectMessage,
  ListDirectMessagesResponse,
} from "#/gen/chat/v1/direct_message_service_pb";

/** サイドバーの行やチャンネルの「その他」メニューから行う、スター・ミュート・既読の操作 */
export const useChannelListActions = (workspaceId: string) => {
  const queryClient = useQueryClient();
  // DM の一覧は取り直さず、変えた DM の行だけを書き換える
  const onSuccess = async (
    channelId: string | undefined,
    update: (dm: DirectMessage) => DirectMessage,
  ) => {
    queryClient.setQueriesData<ListDirectMessagesResponse>(
      { queryKey: dmListKey(workspaceId) },
      (res) =>
        res && {
          ...res,
          directMessages: res.directMessages.map((dm) => (dm.id === channelId ? update(dm) : dm)),
        },
    );
    await queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) });
  };
  const setStarred = useMutation(ChannelService.method.setChannelStarred, {
    onSuccess: (_, { channelId, starred = false }) =>
      onSuccess(channelId, (dm) => ({ ...dm, isStarred: starred })),
  });
  const setMuted = useMutation(ChannelService.method.setChannelMuted, {
    onSuccess: (_, { channelId, muted = false }) =>
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
