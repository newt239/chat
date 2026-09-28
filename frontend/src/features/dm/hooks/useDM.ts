import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";

export const useDMs = (workspaceId: string) =>
  useQuery(
    DirectMessageService.method.listDirectMessages,
    { workspaceId },
    { select: (res) => res.directMessages },
  );

const useInvalidateDMs = () => {
  const queryClient = useQueryClient();

  return async () => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        schema: DirectMessageService.method.listDirectMessages,
      }),
    });
  };
};

export const useCreateDM = () =>
  useMutation(DirectMessageService.method.createDirectMessage, { onSuccess: useInvalidateDMs() });

export const useCreateGroupDM = () =>
  useMutation(DirectMessageService.method.createGroupDirectMessage, {
    onSuccess: useInvalidateDMs(),
  });
