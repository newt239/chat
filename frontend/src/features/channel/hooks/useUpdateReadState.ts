import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { channelListKey } from "#/features/channel/hooks/useChannel";
import { ReadStateService } from "#/gen/chat/v1/read_state_service_pb";

export const useUpdateReadState = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation(ReadStateService.method.updateReadState, {
    onSuccess: async () => {
      // チャンネル一覧を再取得してバッジを更新
      await queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) });
    },
  });
};
