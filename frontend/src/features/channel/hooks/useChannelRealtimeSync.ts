import { useEffect } from "react";

import { useQueryClient } from "@tanstack/react-query";
import { useSetAtom } from "jotai";

import { addChannelPinsDeltaAtom } from "#/providers/store/ui";
import { useWsClient } from "#/providers/ws/useWsClient";

import type { components } from "#/lib/api/schema";

type Channel = components["schemas"]["Channel"];

/** WebSocket イベントからチャンネル一覧の未読バッジとピン件数を更新する。 表示中のチャンネルは既読として扱うため未読を加算しない。 */
export const useChannelRealtimeSync = (
  workspaceId: string | null,
  currentChannelId: string | null,
) => {
  const queryClient = useQueryClient();
  const { wsClient } = useWsClient();
  const addPinsDelta = useSetAtom(addChannelPinsDeltaAtom);

  useEffect(() => {
    if (!wsClient || workspaceId === null) {
      return undefined;
    }

    const queryKey = ["workspaces", workspaceId, "channels"];
    const updateChannel = (channelId: string, update: (channel: Channel) => Channel) => {
      queryClient.setQueryData<Channel[]>(queryKey, (channels) =>
        channels?.map((channel) => (channel.id === channelId ? update(channel) : channel)),
      );
    };

    const unsubscribes = [
      wsClient.on("new_message", ({ channel_id }) => {
        if (channel_id === currentChannelId) {
          return;
        }
        updateChannel(channel_id, (channel) => ({
          ...channel,
          unreadCount: (channel.unreadCount ?? 0) + 1,
        }));
      }),

      wsClient.on("unread_count", ({ channel_id, has_mention, unread_count }) => {
        updateChannel(channel_id, (channel) => ({
          ...channel,
          hasMention: has_mention,
          unreadCount: unread_count,
        }));
      }),

      wsClient.on("pin_created", ({ channel_id }) => {
        addPinsDelta({ channelId: channel_id, delta: 1 });
        void queryClient.invalidateQueries({ queryKey: ["channels", channel_id, "pins"] });
      }),

      wsClient.on("pin_deleted", ({ channel_id }) => {
        addPinsDelta({ channelId: channel_id, delta: -1 });
        void queryClient.invalidateQueries({ queryKey: ["channels", channel_id, "pins"] });
      }),
    ];

    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [wsClient, workspaceId, currentChannelId, queryClient, addPinsDelta]);
};
