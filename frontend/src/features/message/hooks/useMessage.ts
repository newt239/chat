import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { MessagesResponse } from "../types";

import { apiClient } from "@/lib/api/client";


interface CreateMessageInput {
  body: string;
}

export function useMessages(channelId: string | null) {
  return useQuery({
    queryKey: ["channels", channelId, "messages"],
    queryFn: async (): Promise<MessagesResponse> => {
      if (channelId === null) {
        return { messages: [], hasMore: false };
      }

      const { data, error } = await apiClient.GET("/api/channels/{channelId}/messages", {
        params: { path: { channelId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "Failed to fetch messages");
      }

      // API レスポンスを MessagesResponse 型にキャスト
      // バックエンドが user フィールドを含むようになったため
      return data as MessagesResponse;
    },
    enabled: channelId !== null,
  });
}

export function useSendMessage(channelId: string | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateMessageInput) => {
      if (channelId === null) {
        throw new Error("チャンネルが選択されていません");
      }

      const { data, error } = await apiClient.POST("/api/channels/{channelId}/messages", {
        params: { path: { channelId } },
        body: { body: input.body },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "Failed to send message");
      }

      return data;
    },
    onSuccess: async () => {
      if (channelId !== null) {
        await queryClient.invalidateQueries({ queryKey: ["channels", channelId, "messages"] });
      }
    },
  });
}
