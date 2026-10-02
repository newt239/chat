import { useMutation } from "@connectrpc/connect-query";

import { ReactionService } from "#/gen/chat/v1/reaction_service_pb";

// 自分が付けていなければ付け、付けていれば外す。表示は WebSocket の差分で更新する
export const useToggleReaction = (messageId: string) => {
  const addReaction = useMutation(ReactionService.method.addReaction);
  const removeReaction = useMutation(ReactionService.method.removeReaction);

  return (emoji: string, hasReacted: boolean) => {
    (hasReacted ? removeReaction : addReaction).mutate({ emoji, messageId });
  };
};
