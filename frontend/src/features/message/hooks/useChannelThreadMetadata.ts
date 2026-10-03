import { skipToken, useQuery } from "@connectrpc/connect-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";

import type { ThreadMetadata } from "#/gen/chat/v1/message_pb";

/** チャンネル内のメッセージ ID からスレッドメタデータを引けるようにする */
export const useChannelThreadMetadata = (channelId: string | null, includeDescendants: boolean) => {
  const { data: messages } = useQuery(
    MessageService.method.listMessagesWithThread,
    channelId === null ? skipToken : { channelId, includeDescendants },
    { select: (res) => res.messages },
  );

  const map = new Map<string, ThreadMetadata>();
  for (const message of messages ?? []) {
    if (message.threadMetadata) {
      map.set(message.id, message.threadMetadata);
    }
  }
  return map;
};
