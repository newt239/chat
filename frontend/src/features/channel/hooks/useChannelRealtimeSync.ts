import { useEffect } from "react";

import { createConnectQueryKey } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useAtomValue } from "jotai";

import { channelListKey } from "#/features/channel/hooks/useChannel";
import { pinListKey } from "#/features/pin/hooks/usePinActions";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { myUserIdAtom } from "#/providers/store/auth";
import { useWsClient } from "#/providers/ws/useWsClient";

import type { Channel, ListChannelsResponse } from "#/gen/chat/v1/channel_service_pb";

/**
 * WebSocket イベントからチャンネル一覧の未読バッジ・最新メッセージの日時とピン件数を更新する。 表示中のチャンネルは既読として扱うため未読を加算しない。
 * 再接続したら切断中に届かなかった分を取り直す。
 */
export const useChannelRealtimeSync = (
  workspaceId: string | null,
  currentChannelId: string | null,
) => {
  const queryClient = useQueryClient();
  const { wsClient } = useWsClient();
  const currentUserId = useAtomValue(myUserIdAtom);

  useEffect(() => {
    if (!wsClient || workspaceId === null) {
      return undefined;
    }

    const updateChannel = (channelId: string, update: (channel: Channel) => Channel) => {
      // useQuery のキーには transport も含まれるため、完全一致ではなく部分一致で更新する
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

    const unsubscribes = [
      wsClient.onReconnect(() => {
        void queryClient.invalidateQueries({ queryKey: channelListKey(workspaceId) });
        void queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({ cardinality: undefined, schema: MessageService }),
        });
      }),

      wsClient.on("newMessage", ({ channelId, message }) => {
        // 新しいメッセージ順の並びに使う。スレッドの返信は数えない
        const lastMessageAt = message?.parentId === undefined ? message?.createdAt : undefined;
        updateChannel(channelId, (channel) => ({
          ...channel,
          lastMessageAt: lastMessageAt ?? channel.lastMessageAt,
          unreadCount: channel.unreadCount + (channelId === currentChannelId ? 0 : 1),
        }));
      }),

      wsClient.on("unreadCount", ({ channelId, hasMention, mentionCount, unreadCount }) => {
        updateChannel(channelId, (channel) => ({
          ...channel,
          hasMention,
          mentionCount,
          unreadCount,
        }));
      }),

      // 自分の操作は usePinActions で取り直すため二重に取らない
      wsClient.on("pinCreated", ({ channelId, pinnedBy }) => {
        if (pinnedBy !== currentUserId) {
          void queryClient.invalidateQueries({ queryKey: pinListKey(channelId) });
        }
      }),

      wsClient.on("pinDeleted", ({ channelId, pinnedBy }) => {
        if (pinnedBy !== currentUserId) {
          void queryClient.invalidateQueries({ queryKey: pinListKey(channelId) });
        }
      }),
    ];

    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [wsClient, workspaceId, currentChannelId, queryClient, currentUserId]);
};
