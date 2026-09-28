import { useCallback, useEffect, useState } from "react";

import { callUnaryMethod } from "@connectrpc/connect-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { transport } from "#/lib/api/transport";

import { MESSAGES_PAGE_SIZE } from "./useMessage";

import type { TimelineItem } from "#/gen/chat/v1/message_pb";

import type { Timestamp } from "@bufbuild/protobuf/wkt";

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
    async (until: Timestamp | undefined) => {
      if (channelId === null) {
        return;
      }

      setIsLoading(true);
      try {
        const response = await callUnaryMethod(transport, MessageService.method.listMessages, {
          channelId,
          limit: MESSAGES_PAGE_SIZE,
          until,
        });
        setOlderItems((prev) => [...response.messages, ...prev]);
        setHasMore(response.hasMore);
      } finally {
        setIsLoading(false);
      }
    },
    [channelId],
  );

  return { hasMore, isLoading, loadOlder, olderItems };
};
