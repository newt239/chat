import { useMutation, useQueryClient } from "@tanstack/react-query";

//

export type UpdateChannelInput = {
  channelId: string;
  name?: string;
  description?: string | null;
  isPrivate?: boolean;
};

export function useUpdateChannel() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: UpdateChannelInput) => {
      void input; // 型上の引数を使用済みとして扱う
      // 現在のOpenAPIスキーマにチャンネル更新エンドポイントが存在しないため未実装
      // 実装時にはスキーマ追加後にAPI呼び出しへ置換する
      throw new Error("チャンネル更新APIは未実装です");
    },
    onSuccess: async (_data, variables) => {
      await queryClient.invalidateQueries({ queryKey: ["channels", variables.channelId] });
    },
  });
}


