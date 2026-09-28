import { useEffect, useMemo } from "react";

import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useSetAtom } from "jotai";

import { PinService } from "#/gen/chat/v1/pin_service_pb";
import { toDate } from "#/lib/timestamp";
import { setChannelPinsCountAtom } from "#/providers/store/ui";

const PIN_LIMIT = 100;

const usePins = (channelId: string | null) =>
  useQuery(
    PinService.method.listPins,
    channelId === null ? skipToken : { channelId, limit: PIN_LIMIT },
    {
      // メッセージが削除されたピンは表示できないため除く
      select: (res) =>
        res.pins.flatMap(({ message, ...pin }) =>
          message === undefined ? [] : [{ ...pin, message }],
        ),
    },
  );

export const usePinnedMessages = (channelId: string | null) => {
  const setPinsCount = useSetAtom(setChannelPinsCountAtom);
  const query = usePins(channelId);

  useEffect(() => {
    if (channelId && query.data) {
      setPinsCount({ channelId, count: query.data.length });
    }
  }, [channelId, query.data, setPinsCount]);

  const pinsSorted = useMemo(
    () =>
      (query.data ?? []).toSorted(
        (a, b) => toDate(b.pinnedAt).getTime() - toDate(a.pinnedAt).getTime(),
      ),
    [query.data],
  );

  return { ...query, pins: pinsSorted };
};

export const useIsPinned = (messageId: string, channelId: string | null) => {
  const { data: pins } = usePins(channelId);
  return pins?.some((pin) => pin.message.id === messageId) ?? false;
};
