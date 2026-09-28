import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

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

const useInvalidateBookmarks = () => {
  const queryClient = useQueryClient();

  return async () => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        schema: BookmarkService.method.listBookmarks,
      }),
    });
  };
};

export const useAddBookmark = () =>
  useMutation(BookmarkService.method.addBookmark, { onSuccess: useInvalidateBookmarks() });

export const useRemoveBookmark = () =>
  useMutation(BookmarkService.method.removeBookmark, { onSuccess: useInvalidateBookmarks() });

export const useIsBookmarked = (messageId: string) => {
  const { data: bookmarks } = useBookmarks();

  return bookmarks?.some((bookmark) => bookmark.message.id === messageId) ?? false;
};
