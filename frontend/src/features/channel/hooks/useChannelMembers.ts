import { useQuery } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

export const useChannelMembers = (channelId: string | null) =>
  useQuery({
    queryKey: ["channels", channelId, "members"],
    queryFn: async (): Promise<components["schemas"]["ChannelMemberInfo"][]> => {
      if (channelId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/channels/{channelId}/members", {
        params: { path: { channelId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "チャンネルメンバーの取得に失敗しました");
      }

      return data.members;
    },
    enabled: channelId !== null,
  });
