import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { channelListKey } from "#/features/channel/hooks/useChannel";
import { ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";

/** チャンネルメンバーの招待・追放・退出・ロール変更をまとめて提供する */
export const useChannelMemberActions = (workspaceId: string) => {
  const queryClient = useQueryClient();

  const onSuccess = async (_: unknown, { channelId }: { channelId?: string }) => {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          input: { channelId },
          schema: ChannelMemberService.method.listChannelMembers,
        }),
      }),
      queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) }),
    ]);
  };

  const invite = useMutation(ChannelMemberService.method.inviteChannelMember, { onSuccess });
  const remove = useMutation(ChannelMemberService.method.removeChannelMember, { onSuccess });
  const updateRole = useMutation(ChannelMemberService.method.updateChannelMemberRole, {
    onSuccess,
  });
  const join = useMutation(ChannelMemberService.method.joinChannel, { onSuccess });
  const leave = useMutation(ChannelMemberService.method.leaveChannel, { onSuccess });

  return { invite, join, leave, remove, updateRole };
};
