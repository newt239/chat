import { useEffect, useMemo, useRef, useState } from "react";

import { callUnaryMethod } from "@connectrpc/connect-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { transport } from "#/lib/api/transport";

import { MESSAGES_PAGE_SIZE } from "./useMessage";

import type { TimelineItem } from "#/gen/chat/v1/message_pb";
import type { ListMessagesResponse } from "#/gen/chat/v1/message_service_pb";

type Page = { items: TimelineItem[]; hasNext: boolean };
type UseMessagePagesArgs = {
  channelId: string | null;
  includeDescendants: boolean;
  // 変わったら読み込んだ前後を捨てる
  jumpDate: string | null;
  base: ListMessagesResponse | undefined;
};
type Direction = "older" | "newer";

/** 最初に取得した範囲の前後をスクロールに合わせて足していく。項目はどれも新しい順に並べる */
export const useMessagePages = ({
  channelId,
  includeDescendants,
  jumpDate,
  base,
}: UseMessagePagesArgs) => {
  const [older, setOlder] = useState<Page | null>(null);
  const [newer, setNewer] = useState<Page | null>(null);
  const [loading, setLoading] = useState<Direction | null>(null);
  // スクロールのたびに呼ばれるため、読み込み中は重ねて取得しない
  const loadingRef = useRef(false);
  // 切り替える前のチャンネルの応答を捨てる
  const generationRef = useRef(0);

  useEffect(() => {
    generationRef.current += 1;
    loadingRef.current = false;
    setOlder(null);
    setNewer(null);
    setLoading(null);
  }, [channelId, includeDescendants, jumpDate]);

  const baseItems = base?.messages ?? [];
  const olderItems = older?.items ?? [];
  const newerItems = newer?.items ?? [];

  const load = async (direction: Direction) => {
    const boundary =
      direction === "older"
        ? [...baseItems, ...olderItems].at(-1)?.createdAt
        : [...newerItems, ...baseItems].at(0)?.createdAt;
    if (channelId === null || boundary === undefined || loadingRef.current) {
      return;
    }

    const generation = generationRef.current;
    loadingRef.current = true;
    setLoading(direction);
    try {
      const response = await callUnaryMethod(transport, MessageService.method.listMessages, {
        channelId,
        includeDescendants,
        limit: MESSAGES_PAGE_SIZE,
        ...(direction === "older" ? { until: boundary } : { since: boundary }),
      });
      if (generation !== generationRef.current) {
        return;
      }
      if (direction === "older") {
        setOlder({ hasNext: response.hasMore, items: [...olderItems, ...response.messages] });
      } else {
        setNewer({ hasNext: response.hasNewer, items: [...response.messages, ...newerItems] });
      }
    } finally {
      if (generation === generationRef.current) {
        loadingRef.current = false;
        setLoading(null);
      }
    }
  };

  // 同じ内容なら同じ配列を返し、タイムラインを作り直させない
  const items = useMemo(
    () => base && [...(newer?.items ?? []), ...base.messages, ...(older?.items ?? [])],
    [base, newer, older],
  );

  return {
    hasMore: older?.hasNext ?? base?.hasMore ?? false,
    hasNewer: newer?.hasNext ?? base?.hasNewer ?? false,
    items,
    load,
    loading,
  };
};
