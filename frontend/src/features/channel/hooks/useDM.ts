import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";

import type {
  DirectMessage,
  ListDirectMessagesResponse,
} from "#/gen/chat/v1/direct_message_service_pb";

import type { QueryClient } from "@tanstack/react-query";

export const dmListKey = (workspaceId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { workspaceId },
    schema: DirectMessageService.method.listDirectMessages,
  });

// DM の一覧は取り直さず、変えた DM の行だけを書き換える
export const updateListedDM = (
  queryClient: QueryClient,
  { channelId, workspaceId }: { channelId: string; workspaceId: string },
  update: (dm: DirectMessage) => DirectMessage,
) => {
  queryClient.setQueriesData<ListDirectMessagesResponse>(
    { queryKey: dmListKey(workspaceId) },
    (res) =>
      res && {
        ...res,
        directMessages: res.directMessages.map((dm) => (dm.id === channelId ? update(dm) : dm)),
      },
  );
};

export const useDMs = (workspaceId: string) =>
  useQuery(
    DirectMessageService.method.listDirectMessages,
    { workspaceId },
    { select: (res) => res.directMessages },
  );

const useInvalidateDMs = () => {
  const queryClient = useQueryClient();
  return async (_: object, { workspaceId = "" }: { workspaceId?: string }) => {
    await queryClient.invalidateQueries({ queryKey: dmListKey(workspaceId) });
  };
};

export const useCreateDM = () =>
  useMutation(DirectMessageService.method.createDirectMessage, { onSuccess: useInvalidateDMs() });

export const useCreateGroupDM = () =>
  useMutation(DirectMessageService.method.createGroupDirectMessage, {
    onSuccess: useInvalidateDMs(),
  });
