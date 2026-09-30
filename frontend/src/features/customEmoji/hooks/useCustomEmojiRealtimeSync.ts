import { useEffect } from "react";

import { useQueryClient } from "@tanstack/react-query";

import { useWsClient } from "#/providers/ws/useWsClient";

import { customEmojiListKey } from "./useCustomEmojis";

/** 誰かが絵文字を登録・削除したら一覧を取り直す */
export const useCustomEmojiRealtimeSync = (workspaceId: string) => {
  const queryClient = useQueryClient();
  const { wsClient } = useWsClient();

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }
    const refetch = () => {
      void queryClient.invalidateQueries({ queryKey: customEmojiListKey(workspaceId) });
    };
    const unsubscribes = [
      wsClient.on("customEmojiCreated", refetch),
      wsClient.on("customEmojiDeleted", refetch),
      wsClient.onReconnect(refetch),
    ];
    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
    };
  }, [queryClient, wsClient, workspaceId]);
};
