import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "@/lib/api/client";

export const useBookmarks = () => {
  return useQuery({
    queryKey: ["bookmarks"],
    queryFn: async () => {
      const response = await api.GET("/api/bookmarks");
      if (response.error) {
        throw new Error(response.error.error);
      }
      return response.data;
    },
  });
};

export const useAddBookmark = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ messageId }: { messageId: string }) => {
      const response = await api.POST("/api/messages/{messageId}/bookmarks", {
        params: { path: { messageId } },
      });
      if (response.error) {
        throw new Error(response.error.error);
      }
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
    },
  });
};

export const useRemoveBookmark = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ messageId }: { messageId: string }) => {
      const response = await api.DELETE("/api/messages/{messageId}/bookmarks", {
        params: { path: { messageId } },
      });
      if (response.error) {
        throw new Error(response.error.error);
      }
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
    },
  });
};

export const useIsBookmarked = (messageId: string) => {
  const { data: bookmarks } = useBookmarks();

  return bookmarks?.bookmarks.some((bookmark) => bookmark.message?.id === messageId) ?? false;
};
