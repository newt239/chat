import { useMemo } from "react";

import { useQuery } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { ThreadMetadata } from "../types";

/** チャンネル内のメッセージ ID からスレッドメタデータを引けるようにする */
export const useChannelThreadMetadata = (channelId: string | null) => {
  const { data: messages } = useQuery({
    enabled: channelId !== null,
    queryFn: async () => {
      if (channelId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/channels/{channelId}/messages/with-threads", {
        params: { path: { channelId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data.messages;
    },
    queryKey: ["channels", channelId, "messages", "with-threads"],
  });

  return useMemo(() => {
    const map = new Map<string, ThreadMetadata>();
    for (const message of messages ?? []) {
      if (message.threadMetadata) {
        map.set(message.id, message.threadMetadata);
      }
    }
    return map;
  }, [messages]);
};
