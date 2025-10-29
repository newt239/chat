import { useEffect, useMemo } from "react";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useSetAtom } from "jotai";

import { api } from "@/lib/api/client";
import { setChannelPinsCountAtom } from "@/providers/store/ui";

type PinnedMessage = {
  id: string;
  messageId: string;
  channelId: string;
  pinnedAt: string;
  pinnedBy: string;
  // サマリーとして message 本体の最小情報（バックエンドのOpenAPIに準拠）
  message: {
    id: string;
    body: string;
    createdAt: string;
    user?: {
      id: string;
      displayName: string;
      avatarUrl?: string | null;
    } | null;
  } | null;
};

type PinnedListResponse = {
  pins: PinnedMessage[];
  nextCursor?: string | null;
};

export function usePinnedMessages(channelId: string | null, limit = 100) {
  const setPinsCount = useSetAtom(setChannelPinsCountAtom);

  const query = useQuery({
    queryKey: ["channels", channelId, "pins"],
    enabled: channelId !== null,
    queryFn: async () => {
      if (channelId === null) return { pins: [], nextCursor: null } as PinnedListResponse;

      // OpenAPI スキーマ更新前の一時対応
      // @ts-expect-error OpenAPI スキーマに pins が未反映
      const { data, error } = await api.GET("/api/channels/{channelId}/pins", {
        params: { path: { channelId }, query: { limit } },
      });

      if (error || !data) {
        // @ts-expect-error error 型はスキーマ反映後に解消
        throw new Error(error?.error ?? "ピン一覧の取得に失敗しました");
      }

      // API スキーマ準拠前提。必要であれば zod 検証を追加
      return data as unknown as PinnedListResponse;
    },
  });

  useEffect(() => {
    if (channelId && query.data) {
      setPinsCount({ channelId, count: query.data.pins.length });
    }
  }, [channelId, query.data, setPinsCount]);

  const pinsSorted = useMemo(() => {
    const pins = query.data?.pins ?? [];
    return [...pins].sort(
      (a, b) => new Date(b.pinnedAt).getTime() - new Date(a.pinnedAt).getTime()
    );
  }, [query.data]);

  return { ...query, pins: pinsSorted };
}

export function useIsPinned(messageId: string | null, channelId: string | null) {
  const queryClient = useQueryClient();
  const data = queryClient.getQueryData<PinnedListResponse>(["channels", channelId, "pins"]);
  if (!messageId || !channelId || !data) return false;
  return data.pins.some((p) => p.message?.id === messageId);
}