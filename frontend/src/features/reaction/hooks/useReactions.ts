import { useMutation } from "@connectrpc/connect-query";

import { useInvalidateMessages } from "#/features/message/hooks/useMessage";
import { ReactionService } from "#/gen/chat/v1/reaction_service_pb";

const useAddReaction = () =>
  useMutation(ReactionService.method.addReaction, { onSuccess: useInvalidateMessages() });

const useRemoveReaction = () =>
  useMutation(ReactionService.method.removeReaction, { onSuccess: useInvalidateMessages() });

// 自分が付けていなければ付け、付けていれば外す
export const useToggleReaction = (messageId: string) => {
  const addReaction = useAddReaction();
  const removeReaction = useRemoveReaction();

  return (emoji: string, hasReacted: boolean) => {
    (hasReacted ? removeReaction : addReaction).mutate({ emoji, messageId });
  };
};
