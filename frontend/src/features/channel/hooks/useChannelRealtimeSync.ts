import { useEffect } from "react";

import { createConnectQueryKey } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { useWsClient } from "#/providers/ws/useWsClient";

import { channelListKey } from "./useChannel";
import { dmListKey } from "./useDM";

import type { Channel, ListChannelsResponse } from "#/gen/chat/v1/channel_service_pb";
import type {
  DirectMessage,
  ListDirectMessagesResponse,
} from "#/gen/chat/v1/direct_message_service_pb";

/** WebSocket イベントでチャンネルと DM の一覧の未読を更新する。表示中は既読扱いにし、再接続したら取り直す */
export const useChannelRealtimeSync = (workspaceId: string, currentChannelId: string | null) => {
  const queryClient = useQueryClient();
  const wsClient = useWsClient();

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }

    // useQuery のキーには transport も含まれるため、完全一致ではなく部分一致で更新する
    const updateChannel = (channelId: string, update: (channel: Channel) => Channel) => {
      queryClient.setQueriesData<ListChannelsResponse>(
        { queryKey: channelListKey(workspaceId) },
        (res) =>
          res && {
            ...res,
            channels: res.channels.map((channel) =>
              channel.id === channelId ? update(channel) : channel,
            ),
          },
      );
    };
    const updateDM = (channelId: string, update: (dm: DirectMessage) => DirectMessage) => {
      queryClient.setQueriesData<ListDirectMessagesResponse>(
        { queryKey: dmListKey(workspaceId) },
        (res) =>
          res && {
            ...res,
            directMessages: res.directMessages.map((dm) => (dm.id === channelId ? update(dm) : dm)),
          },
      );
    };

    const unsubscribes = [
      wsClient.onReconnect(() => {
        void queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) });
        void queryClient.invalidateQueries({ queryKey: dmListKey(workspaceId) });
        void queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({ cardinality: undefined, schema: MessageService }),
        });
      }),

      wsClient.on("newMessage", ({ channelId, message }) => {
        const unread = channelId === currentChannelId ? 0 : 1;
        // 新しいメッセージ順の並びに使う。スレッドの返信は数えない
        const lastMessageAt = message?.parentId === undefined ? message?.createdAt : undefined;
        updateChannel(channelId, (channel) => ({
          ...channel,
          lastMessageAt: lastMessageAt ?? channel.lastMessageAt,
          unreadCount: channel.unreadCount + unread,
        }));
        updateDM(channelId, (dm) => ({ ...dm, unreadCount: dm.unreadCount + unread }));
      }),

      wsClient.on("unreadCount", ({ channelId, hasMention, mentionCount, unreadCount }) => {
        updateChannel(channelId, (channel) => ({
          ...channel,
          hasMention,
          mentionCount,
          unreadCount,
        }));
        updateDM(channelId, (dm) => ({ ...dm, hasMention, unreadCount }));
      }),
    ];

    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [wsClient, workspaceId, currentChannelId, queryClient]);
};
