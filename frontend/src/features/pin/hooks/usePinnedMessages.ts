import { useMemo } from "react";

import { useQuery } from "@connectrpc/connect-query";

import { PinService } from "#/gen/chat/v1/pin_service_pb";
import { toDate } from "#/lib/timestamp";

import type { ListPinsResponse } from "#/gen/chat/v1/pin_service_pb";

const PIN_LIMIT = 100;

// メッセージが削除されたピンは表示できないため除く
const toVisiblePins = (res: ListPinsResponse) =>
  res.pins.flatMap(({ message, ...pin }) => (message === undefined ? [] : [{ ...pin, message }]));

/** ヘッダーのバッジに出す件数。ピンの操作や WebSocket のイベントで一覧ごと invalidate して更新する */
export const usePinCount = (channelId: string) =>
  useQuery(
    PinService.method.listPins,
    { channelId, limit: PIN_LIMIT },
    { select: (res) => toVisiblePins(res).length },
  ).data ?? 0;

export const usePinnedMessages = (channelId: string) => {
  const query = useQuery(
    PinService.method.listPins,
    { channelId, limit: PIN_LIMIT },
    { select: toVisiblePins },
  );

  const pinsSorted = useMemo(
    () =>
      (query.data ?? []).toSorted(
        (a, b) => toDate(b.pinnedAt).getTime() - toDate(a.pinnedAt).getTime(),
      ),
    [query.data],
  );

  return { ...query, pins: pinsSorted };
};
