import { useCallback, useEffect, useRef } from "react";

import { useWsClient } from "#/providers/ws/useWsClient";

const TYPING_THROTTLE_MS = 2_000;
const TYPING_IDLE_MS = 5_000;

/** 入力中であることを WebSocket で通知する。連投を抑えるため間引いて送信する */
export const useTypingNotifier = (channelId: string) => {
  const { wsClient } = useWsClient();
  const lastSentAtRef = useRef(0);
  const idleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearIdleTimer = useCallback(() => {
    if (idleTimerRef.current !== null) {
      clearTimeout(idleTimerRef.current);
      idleTimerRef.current = null;
    }
  }, []);

  useEffect(() => clearIdleTimer, [clearIdleTimer]);

  const notifyTyping = useCallback(() => {
    if (!wsClient) {
      return;
    }

    const now = Date.now();
    if (now - lastSentAtRef.current > TYPING_THROTTLE_MS) {
      lastSentAtRef.current = now;
      wsClient.typing(channelId);
    }

    clearIdleTimer();
    idleTimerRef.current = setTimeout(() => {
      lastSentAtRef.current = 0;
      wsClient.stopTyping(channelId);
    }, TYPING_IDLE_MS);
  }, [wsClient, channelId, clearIdleTimer]);

  const notifyStopTyping = useCallback(() => {
    clearIdleTimer();
    lastSentAtRef.current = 0;
    wsClient?.stopTyping(channelId);
  }, [wsClient, channelId, clearIdleTimer]);

  return { notifyStopTyping, notifyTyping };
};
