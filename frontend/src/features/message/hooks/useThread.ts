import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { threadRepliesResponseSchema } from "../schemas";

import { api } from "@/lib/api/client";

type CreateThreadReplyInput = {
  body: string;
};

/**
 * スレッドの返信一覧を取得するフック
 */
export function useThreadReplies(messageId: string | null) {
  return useQuery({
    queryKey: ["messages", messageId, "thread", "replies"],
    queryFn: async () => {
      if (messageId === null) {
        return null;
      }

      const { data, error } = await api.GET("/api/messages/{messageId}/thread", {
        params: { path: { messageId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "スレッド返信の取得に失敗しました");
      }

      const parsed = threadRepliesResponseSchema.safeParse(data);

      if (!parsed.success) {
        console.error("スレッド返信取得のスキーマ検証エラー:", parsed.error);
        console.error("受信したデータ:", JSON.stringify(data, null, 2));
        throw new Error("スレッド返信取得のレスポンス形式が想定と異なります");
      }

      return parsed.data;
    },
    enabled: messageId !== null,
  });
}

/**
 * スレッドに返信を送信するフック
 */
export function useSendThreadReply(messageId: string | null, channelId: string | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateThreadReplyInput) => {
      if (messageId === null) {
        throw new Error("親メッセージが選択されていません");
      }

      if (channelId === null) {
        throw new Error("チャンネルが選択されていません");
      }

      const { data, error } = await api.POST("/api/channels/{channelId}/messages", {
        params: { path: { channelId } },
        body: {
          body: input.body,
          parentId: messageId,
        },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "スレッドへの返信送信に失敗しました");
      }

      return data;
    },
    onSuccess: async () => {
      if (messageId !== null && channelId !== null) {
        // スレッド返信一覧を再取得
        await queryClient.invalidateQueries({
          queryKey: ["messages", messageId, "thread", "replies"],
        });
        // スレッドメタデータを再取得
        await queryClient.invalidateQueries({
          queryKey: ["messages", messageId, "thread", "metadata"],
        });
        // メッセージ一覧も再取得（スレッドプレビューの更新のため）
        await queryClient.invalidateQueries({
          queryKey: ["channels", channelId, "messages"],
        });
      }
    },
  });
}
