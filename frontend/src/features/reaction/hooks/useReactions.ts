import { useMutation } from "@connectrpc/connect-query";
import { useAtomValue } from "jotai";

import { useInvalidateMessageLists } from "#/features/message/hooks/useInvalidateMessageLists";
import { ReactionService } from "#/gen/chat/v1/reaction_service_pb";
import { myUserIdAtom } from "#/providers/store/auth";

import type { Message } from "#/gen/chat/v1/message_pb";

// 自分が付けていなければ付け、付けていれば外す。チャンネルの表示は WebSocket の差分で更新する
export const useToggleReaction = (message: Message) => {
  const myId = useAtomValue(myUserIdAtom);
  const invalidateMessageLists = useInvalidateMessageLists();
  const addReaction = useMutation(ReactionService.method.addReaction, {
    onSuccess: invalidateMessageLists,
  });
  const removeReaction = useMutation(ReactionService.method.removeReaction, {
    onSuccess: invalidateMessageLists,
  });

  return (emoji: string) => {
    const hasReacted = message.reactions.some(
      (reaction) => reaction.emoji === emoji && reaction.user?.id === myId,
    );
    (hasReacted ? removeReaction : addReaction).mutate({ emoji, messageId: message.id });
  };
};
