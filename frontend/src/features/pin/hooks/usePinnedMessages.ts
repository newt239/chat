import { useQuery } from "@connectrpc/connect-query";

import { PinService } from "#/gen/chat/v1/pin_service_pb";

/** ヘッダーのバッジに出す件数。ピンの操作や WebSocket のイベントで一覧ごと invalidate して更新する */
export const usePinCount = (channelId: string) =>
  useQuery(PinService.method.listPins, { channelId }, { select: (res) => res.messages.length })
    .data ?? 0;

export const usePinnedMessages = (channelId: string) => {
  const query = useQuery(PinService.method.listPins, { channelId });
  return { ...query, pins: query.data?.messages ?? [] };
};
