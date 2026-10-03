import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { channelListKey } from "#/features/channel/hooks/useChannel";
import { ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { toastError } from "#/lib/toastError";

/** チャンネルメンバーの招待・追放・退出・ロール変更をまとめて提供する */
export const useChannelMemberActions = (workspaceId: string) => {
  const queryClient = useQueryClient();

  const onSuccess = async (_: object, { channelId }: { channelId?: string }) => {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          input: { channelId },
          schema: ChannelMemberService.method.listChannelMembers,
        }),
      }),
      queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) }),
      // 参加状態はブラウズ一覧とプレビュー中のチャンネルにも出ている
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          schema: ChannelService.method.listBrowsableChannels,
        }),
      }),
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          schema: ChannelService.method.searchBrowsableChannels,
        }),
      }),
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          input: { channelId },
          schema: ChannelService.method.getChannel,
        }),
      }),
    ]);
  };

  const invite = useMutation(ChannelMemberService.method.inviteChannelMember, { onSuccess });
  const remove = useMutation(ChannelMemberService.method.removeChannelMember, {
    onError: toastError,
    onSuccess,
  });
  const updateRole = useMutation(ChannelMemberService.method.updateChannelMemberRole, {
    onError: toastError,
    onSuccess,
  });
  const join = useMutation(ChannelMemberService.method.joinChannel, {
    onError: toastError,
    onSuccess,
  });
  const leave = useMutation(ChannelMemberService.method.leaveChannel, {
    onError: toastError,
    onSuccess,
  });

  return { invite, join, leave, remove, updateRole };
};
