import { useEffect, useCallback } from "react";

import { useAtomValue } from "jotai";

import { useWebSocket } from "./WebSocketProvider";

import { currentChannelIdAtom } from "@/providers/store/workspace";

export const useChannelEvents = () => {
  const { client } = useWebSocket();
  const currentChannelId = useAtomValue(currentChannelIdAtom);

  const joinChannel = useCallback(
    (channelId: string) => {
      if (client) {
        client.send("join_channel", { channel_id: channelId });
      }
    },
    [client]
  );

  const leaveChannel = useCallback(
    (channelId: string) => {
      if (client) {
        client.send("leave_channel", { channel_id: channelId });
      }
    },
    [client]
  );

  // チャンネルが変更されたときに自動的に参加・退出
  useEffect(() => {
    if (!client || !currentChannelId) return;

    // 現在のチャンネルに参加
    joinChannel(currentChannelId);

    return () => {
      // コンポーネントがアンマウントされる際にチャンネルから退出
      leaveChannel(currentChannelId);
    };
  }, [client, currentChannelId, joinChannel, leaveChannel]);

  return {
    joinChannel,
    leaveChannel,
  };
};
