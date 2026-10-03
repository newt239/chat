import { useEffect } from "react";

import { createConnectQueryKey } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { useWsClient } from "#/providers/ws/useWsClient";

import type {
  DirectMessage,
  ListDirectMessagesResponse,
} from "#/gen/chat/v1/direct_message_service_pb";

/** WebSocket イベントから DM 一覧の未読数を更新する。表示中の DM は既読として扱う */
export const useDMRealtimeSync = (workspaceId: string, currentChannelId: string | null) => {
  const queryClient = useQueryClient();
  const wsClient = useWsClient();

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }

    const queryKey = createConnectQueryKey({
      cardinality: "finite",
      input: { workspaceId },
      schema: DirectMessageService.method.listDirectMessages,
    });
    const updateDM = (channelId: string, update: (dm: DirectMessage) => DirectMessage) => {
      queryClient.setQueriesData<ListDirectMessagesResponse>(
        { queryKey },
        (res) =>
          res && {
            ...res,
            directMessages: res.directMessages.map((dm) => (dm.id === channelId ? update(dm) : dm)),
          },
      );
    };

    const unsubscribes = [
      wsClient.on("newMessage", ({ channelId }) => {
        if (channelId !== currentChannelId) {
          updateDM(channelId, (dm) => ({ ...dm, unreadCount: dm.unreadCount + 1 }));
        }
      }),
      wsClient.on("unreadCount", ({ channelId, hasMention, unreadCount }) => {
        updateDM(channelId, (dm) => ({ ...dm, hasMention, unreadCount }));
      }),
      // 切断中に届いた DM の未読は差分で追えないため取り直す
      wsClient.onReconnect(() => {
        void queryClient.invalidateQueries({ queryKey });
      }),
    ];
    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [wsClient, queryClient, workspaceId, currentChannelId]);
};
