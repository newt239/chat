import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { DraftService } from "#/gen/chat/v1/draft_service_pb";
import { toastError } from "#/lib/toastError";

export const useDrafts = (workspaceId: string) =>
  useQuery(DraftService.method.listDrafts, { workspaceId }, { select: (res) => res.drafts });

/** 下書きの一覧と、各チャンネル・スレッドの下書きを取り直す */
export const useInvalidateDrafts = () => {
  const queryClient = useQueryClient();
  return () =>
    queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({ cardinality: "finite", schema: DraftService }),
    });
};

export const useDeleteDraft = () =>
  useMutation(DraftService.method.deleteDraft, {
    onError: toastError,
    onSuccess: useInvalidateDrafts(),
  });
