import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { BookmarkService } from "#/gen/chat/v1/bookmark_service_pb";

export const useBookmarks = () =>
  useQuery(
    BookmarkService.method.listBookmarks,
    {},
    {
      // メッセージが削除されたブックマークは表示できないため除く
      select: (res) =>
        res.bookmarks.flatMap(({ message, ...bookmark }) =>
          message === undefined ? [] : [{ ...bookmark, message }],
        ),
    },
  );

// 付いていなければ付け、付いていれば外す
export const useToggleBookmark = (messageId: string) => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { data: bookmarks } = useBookmarks();
  const onSuccess = () =>
    queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        schema: BookmarkService.method.listBookmarks,
      }),
    });
  const addBookmark = useMutation(BookmarkService.method.addBookmark, { onSuccess });
  const removeBookmark = useMutation(BookmarkService.method.removeBookmark, { onSuccess });
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
