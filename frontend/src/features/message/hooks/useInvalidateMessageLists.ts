import { createConnectQueryKey } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { MentionService } from "#/gen/chat/v1/mention_service_pb";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

const listKeys = [
  createConnectQueryKey({ cardinality: "infinite", schema: MentionService.method.listMentions }),
  createConnectQueryKey({
    cardinality: "infinite",
    schema: ThreadService.method.listParticipatingThreads,
  }),
];

// メンション・スレッドの一覧はチャンネルに参加しておらず WebSocket の差分が届かないため、操作したら取り直す
export const useInvalidateMessageLists = () => {
  const queryClient = useQueryClient();
  return () => {
    for (const queryKey of listKeys) {
      void queryClient.invalidateQueries({ queryKey });
    }
  };
};
