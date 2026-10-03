import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { useInvalidateMessageLists } from "#/features/message/hooks/useInvalidateMessageLists";
import { PinService } from "#/gen/chat/v1/pin_service_pb";

import type { Message } from "#/gen/chat/v1/message_pb";

export const pinListKey = (channelId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { channelId },
    schema: PinService.method.listPins,
  });

// メッセージ側のピン留めの表示は WebSocket の差分で更新する
export const useTogglePin = (message: Message) => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const invalidateMessageLists = useInvalidateMessageLists();
  const onSuccess = () => {
    invalidateMessageLists();
    void queryClient.invalidateQueries({ queryKey: pinListKey(message.channelId) });
  };
  const pin = useMutation(PinService.method.createPin, { onSuccess });
  const unpin = useMutation(PinService.method.deletePin, { onSuccess });
  const isPinned = message.pin !== undefined;

  return () => {
    (isPinned ? unpin : pin).mutate(
      { channelId: message.channelId, messageId: message.id },
      {
        onSuccess: () => {
          toast(t(isPinned ? "pin.unpinned" : "pin.pinned"));
        },
      },
    );
  };
};
