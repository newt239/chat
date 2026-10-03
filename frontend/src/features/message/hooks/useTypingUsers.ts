import { useEffect, useState } from "react";

import { useWsClient } from "#/providers/ws/useWsClient";

// 送信側は入力中に 2 秒おきに通知し、5 秒止まると終了を送る。終了が届かなくても消えるよう少し長く待つ
const TYPING_TIMEOUT_MS = 6_000;

export const useTypingUsers = (channelId: string) => {
  const wsClient = useWsClient();
  const [typingUserIds, setTypingUserIds] = useState<string[]>([]);

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }
    const timers = new Map<string, ReturnType<typeof setTimeout>>();
    const remove = (userId: string) => {
      clearTimeout(timers.get(userId));
      timers.delete(userId);
      setTypingUserIds((prev) => prev.filter((id) => id !== userId));
    };
    const unsubscribes = [
      wsClient.on("typing", ({ channelId: eventChannelId, userId }) => {
        if (eventChannelId !== channelId) {
          return;
        }
        clearTimeout(timers.get(userId));
        timers.set(
          userId,
          setTimeout(() => {
            remove(userId);
          }, TYPING_TIMEOUT_MS),
        );
        setTypingUserIds((prev) => (prev.includes(userId) ? prev : [...prev, userId]));
      }),
      wsClient.on("stopTyping", ({ channelId: eventChannelId, userId }) => {
        if (eventChannelId === channelId) {
          remove(userId);
        }
      }),
    ];
    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
      for (const timer of timers.values()) {
        clearTimeout(timer);
      }
      setTypingUserIds([]);
    };
  }, [wsClient, channelId]);

  return typingUserIds;
};
