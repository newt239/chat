import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { AppService } from "#/gen/chat/v1/app_service_pb";

export const useApps = (workspaceId: string) =>
  useQuery(AppService.method.listApps, { workspaceId });

export const useChannelApps = (channelId: string) =>
  useQuery(AppService.method.listChannelApps, { channelId });

/** アプリの作成・編集・URL の再発行・削除とチャンネルへの追加。成功したら一覧を取り直す */
export const useAppActions = () => {
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({ cardinality: "finite", schema: AppService }),
    });
  };
  return {
    addToChannel: useMutation(AppService.method.addAppToChannel, { onSuccess }),
    create: useMutation(AppService.method.createApp, { onSuccess }),
    regenerate: useMutation(AppService.method.regenerateAppToken),
    remove: useMutation(AppService.method.deleteApp, { onSuccess }),
    removeFromChannel: useMutation(AppService.method.removeAppFromChannel, { onSuccess }),
    update: useMutation(AppService.method.updateApp, { onSuccess }),
  };
};
