import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { PinService } from "#/gen/chat/v1/pin_service_pb";

export const pinListKey = (channelId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { channelId },
    schema: PinService.method.listPins,
  });

export const usePinActions = () => {
  const queryClient = useQueryClient();

  // メッセージ側のピン留めの表示は WebSocket の差分で更新する
  const onSuccess = (_: object, { channelId = "" }) =>
    queryClient.invalidateQueries({ queryKey: pinListKey(channelId) });

  const pin = useMutation(PinService.method.createPin, { onSuccess });
  const unpin = useMutation(PinService.method.deletePin, { onSuccess });

  return { pin, unpin };
};
