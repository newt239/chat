import { useCallback, useEffect, useState } from "react";

import { api } from "#/lib/api/client";

import { MESSAGES_PAGE_SIZE } from "./useMessage";

import type { TimelineItem } from "../types";

/** 「さらに読み込む」で取得した過去メッセージを保持する */
export const useOlderMessages = (channelId: string | null, initialHasMore: boolean) => {
  const [olderItems, setOlderItems] = useState<TimelineItem[]>([]);
  const [hasMore, setHasMore] = useState(initialHasMore);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    setOlderItems([]);
    setIsLoading(false);
  }, [channelId]);

  useEffect(() => {
    setHasMore(initialHasMore);
  }, [initialHasMore, channelId]);

  const loadOlder = useCallback(
    async (until: string) => {
      if (channelId === null) {
        return;
      }

      setIsLoading(true);
      try {
        const { data, error } = await api.GET("/api/channels/{channelId}/messages", {
          params: { path: { channelId }, query: { limit: MESSAGES_PAGE_SIZE, until } },
        });

        if (error) {
          throw new Error(error.error);
        }

        setOlderItems((prev) => [...data.messages, ...prev]);
        setHasMore(data.hasMore);
      } finally {
        setIsLoading(false);
      }
    },
    [channelId],
  );

  return { hasMore, isLoading, loadOlder, olderItems };
};
