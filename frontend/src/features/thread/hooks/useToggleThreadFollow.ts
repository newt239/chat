import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { ThreadService } from "#/gen/chat/v1/thread_service_pb";

import { useUpdateListedThread } from "./useParticipatingThreads";

import type { ListMessagesWithThreadResponse } from "#/gen/chat/v1/message_service_pb";
import type { GetThreadMetadataResponse } from "#/gen/chat/v1/thread_service_pb";

/** スレッドのフォローを切り替え、メタデータ・タイムライン・参加中の一覧のキャッシュに反映する */
export const useToggleThreadFollow = (threadId: string) => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const updateListedThread = useUpdateListedThread();
  const follow = useMutation(ThreadService.method.followThread);
  const unfollow = useMutation(ThreadService.method.unfollowThread);

  const setFollowing = (isFollowing: boolean) => {
    (isFollowing ? follow : unfollow).mutate(
      { messageId: threadId },
      {
        onSuccess: () => {
          queryClient.setQueriesData<GetThreadMetadataResponse>(
            {
              queryKey: createConnectQueryKey({
                cardinality: "finite",
                input: { messageId: threadId },
                schema: ThreadService.method.getThreadMetadata,
              }),
            },
            (res) => res?.metadata && { ...res, metadata: { ...res.metadata, isFollowing } },
          );
          queryClient.setQueriesData<ListMessagesWithThreadResponse>(
            {
              queryKey: createConnectQueryKey({
                cardinality: "finite",
                schema: MessageService.method.listMessagesWithThread,
              }),
            },
            (res) =>
              res && {
                ...res,
                messages: res.messages.map((message) =>
                  message.id === threadId && message.threadMetadata
                    ? { ...message, threadMetadata: { ...message.threadMetadata, isFollowing } }
                    : message,
                ),
              },
          );
          updateListedThread(threadId, (thread) => ({ ...thread, isFollowing }));
          toast(t(isFollowing ? "thread.follow.followed" : "thread.follow.unfollowed"));
        },
      },
    );
  };

  return { isPending: follow.isPending || unfollow.isPending, setFollowing };
};
