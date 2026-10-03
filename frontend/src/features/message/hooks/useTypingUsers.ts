import { useEffect, useState } from "react";

import { useWsClient } from "#/providers/ws/useWsClient";

export const useTypingUsers = (channelId: string) => {
  const wsClient = useWsClient();
  const [typingUserIds, setTypingUserIds] = useState<string[]>([]);

  useEffect(() => {
    setTypingUserIds([]);
    if (!wsClient) {
      return undefined;
    }
    const unsubscribes = [
      wsClient.on("typing", ({ channelId: eventChannelId, userId }) => {
        if (eventChannelId === channelId) {
          setTypingUserIds((prev) => (prev.includes(userId) ? prev : [...prev, userId]));
        }
      }),
      wsClient.on("stopTyping", ({ channelId: eventChannelId, userId }) => {
        if (eventChannelId === channelId) {
          setTypingUserIds((prev) => prev.filter((id) => id !== userId));
        }
      }),
    ];
    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [wsClient, channelId]);

  return typingUserIds;
};
