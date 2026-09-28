import { useMutation } from "@connectrpc/connect-query";

import { useInvalidateMessages } from "#/features/message/hooks/useMessage";
import { ReactionService } from "#/gen/chat/v1/reaction_service_pb";

export const useAddReaction = () =>
  useMutation(ReactionService.method.addReaction, { onSuccess: useInvalidateMessages() });

export const useRemoveReaction = () =>
  useMutation(ReactionService.method.removeReaction, { onSuccess: useInvalidateMessages() });
