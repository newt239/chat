import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useSetAtom } from "jotai";

import { PinService } from "#/gen/chat/v1/pin_service_pb";
import { addChannelPinsDeltaAtom } from "#/providers/store/ui";

export const pinListKey = (channelId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { channelId },
    schema: PinService.method.listPins,
  });

export const usePinActions = () => {
  const queryClient = useQueryClient();
  const addPinsDelta = useSetAtom(addChannelPinsDeltaAtom);

  const onSuccess =
    (delta: number) =>
    async (_: unknown, { channelId = "" }) => {
      addPinsDelta({ channelId, delta });
      await queryClient.invalidateQueries({ queryKey: pinListKey(channelId) });
    };

  const pin = useMutation(PinService.method.createPin, { onSuccess: onSuccess(1) });
  const unpin = useMutation(PinService.method.deletePin, { onSuccess: onSuccess(-1) });

  return { pin, unpin };
};
