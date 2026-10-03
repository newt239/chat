import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { BookmarkService } from "#/gen/chat/v1/bookmark_service_pb";
import { toastError } from "#/lib/toastError";

export const useBookmarks = (workspaceId: string) =>
  useQuery(
    BookmarkService.method.listBookmarks,
    { workspaceId },
    {
      // サーバーは常に message を入れるが、proto の型では省略可能なため絞り込む
      select: (res) =>
        res.bookmarks.flatMap(({ message, ...bookmark }) =>
          message === undefined ? [] : [{ ...bookmark, message }],
        ),
    },
  );

// 付いていなければ付け、付いていれば外す
export const useToggleBookmark = (messageId: string, workspaceId: string) => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { data: bookmarks } = useBookmarks(workspaceId);
  const onSuccess = () =>
    queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        schema: BookmarkService.method.listBookmarks,
      }),
    });
  const addBookmark = useMutation(BookmarkService.method.addBookmark, {
    onError: toastError,
    onSuccess,
  });
  const removeBookmark = useMutation(BookmarkService.method.removeBookmark, {
    onError: toastError,
    onSuccess,
  });
  const isBookmarked = bookmarks?.some((bookmark) => bookmark.message.id === messageId) ?? false;

  const toggleBookmark = () => {
    (isBookmarked ? removeBookmark : addBookmark).mutate(
      { messageId },
      {
        onSuccess: () => {
          toast(t(isBookmarked ? "bookmark.removed" : "bookmark.added"));
        },
      },
    );
  };

  return { isBookmarked, toggleBookmark };
};
