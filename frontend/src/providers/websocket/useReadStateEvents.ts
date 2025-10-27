import { useCallback } from "react";

import { useWebSocket } from "./WebSocketProvider";

export const useReadStateEvents = () => {
  const { client } = useWebSocket();

  const updateReadState = useCallback(
    (channelId: string, messageId?: string) => {
      if (client) {
        client.send("update_read_state", {
          channel_id: channelId,
          message_id: messageId,
        });
      }
    },
    [client]
  );

  return {
    updateReadState,
  };
};
