import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { messagesTimelineResponseSchema } from "../schemas";

import { api } from "@/lib/api/client";

type CreateMessageInput = {
  body: string;
  attachmentIds?: string[];
};

type UpdateMessageInput = {
  messageId: string;
  body: string;
};

type DeleteMessageInput = {
  messageId: string;
};

export function useMessages(channelId: string | null) {
  return useQuery({
    queryKey: ["channels", channelId, "messages"],
    queryFn: async () => {
      if (channelId === null) {
        return { messages: [], hasMore: false } as const;
      }

      const { data, error } = await api.GET("/api/channels/{channelId}/messages", {
        params: { path: { channelId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "メッセージ一覧の取得に失敗しました");
      }

      const parsed = messagesTimelineResponseSchema.safeParse(data);

      if (!parsed.success) {
        console.error("メッセージ取得のスキーマ検証エラー:", parsed.error);
        console.error("受信したデータ:", JSON.stringify(data, null, 2));
        throw new Error("メッセージ取得のレスポンス形式が想定と異なります");
      }

      return parsed.data;
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

      const { data, error } = await api.POST("/api/channels/{channelId}/messages", {
        params: { path: { channelId } },
        body: { body: input.body, attachmentIds: input.attachmentIds },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "メッセージの送信に失敗しました");
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

export function useUpdateMessage(channelId: string | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateMessageInput) => {
      const { data, error } = await api.PATCH("/api/messages/{messageId}", {
        params: { path: { messageId: input.messageId } },
        body: { body: input.body },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "メッセージの更新に失敗しました");
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

export function useDeleteMessage(channelId: string | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: DeleteMessageInput) => {
      const { error } = await api.DELETE("/api/messages/{messageId}", {
        params: { path: { messageId: input.messageId } },
      });

      if (error) {
        throw new Error(error?.error ?? "メッセージの削除に失敗しました");
      }
    },
    onSuccess: async () => {
      if (channelId !== null) {
        await queryClient.invalidateQueries({ queryKey: ["channels", channelId, "messages"] });
      }
    },
  });
}
