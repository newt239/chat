import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "@/lib/api/client";

export const useAddReaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ messageId, emoji }: { messageId: string; emoji: string }) => {
      const { error } = await api.POST("/api/messages/{messageId}/reactions", {
        params: { path: { messageId } },
        body: { emoji },
      });
      if (error) throw error;
    },
    onSuccess: (_, { messageId }) => {
      // リアクションとメッセージのクエリを無効化
      queryClient.invalidateQueries({ queryKey: ["reactions", messageId] });
      queryClient.invalidateQueries({ queryKey: ["channels"] });
    },
  });
};

export const useRemoveReaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ messageId, emoji }: { messageId: string; emoji: string }) => {
      const { error } = await api.DELETE("/api/messages/{messageId}/reactions/{emoji}", {
        params: { path: { messageId, emoji } },
      });
      if (error) throw error;
    },
    onSuccess: (_, { messageId }) => {
      // リアクションとメッセージのクエリを無効化
      queryClient.invalidateQueries({ queryKey: ["reactions", messageId] });
      queryClient.invalidateQueries({ queryKey: ["channels"] });
    },
  });
};
