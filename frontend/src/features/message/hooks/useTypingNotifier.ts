import { useEffect, useRef } from "react";
import type { RefObject } from "react";

import { useWsClient } from "#/providers/ws/useWsClient";

const TYPING_THROTTLE_MS = 2_000;
const TYPING_IDLE_MS = 5_000;

const clearTimer = (timerRef: RefObject<ReturnType<typeof setTimeout> | null>) => {
  if (timerRef.current !== null) {
    clearTimeout(timerRef.current);
    timerRef.current = null;
  }
};

/** 入力中であることを WebSocket で通知する。連投を抑えるため間引いて送信する */
export const useTypingNotifier = (channelId: string) => {
  const { wsClient } = useWsClient();
  const lastSentAtRef = useRef(0);
  const idleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      clearTimer(idleTimerRef);
    },
    [],
  );

  const notifyTyping = () => {
    if (!wsClient) {
      return;
    }

    const now = Date.now();
    if (now - lastSentAtRef.current > TYPING_THROTTLE_MS) {
      lastSentAtRef.current = now;
      wsClient.typing(channelId);
    }

    clearTimer(idleTimerRef);
    idleTimerRef.current = setTimeout(() => {
      lastSentAtRef.current = 0;
      wsClient.stopTyping(channelId);
    }, TYPING_IDLE_MS);
  };

  const notifyStopTyping = () => {
    clearTimer(idleTimerRef);
    lastSentAtRef.current = 0;
    wsClient?.stopTyping(channelId);
  };

  return { notifyStopTyping, notifyTyping };
};
