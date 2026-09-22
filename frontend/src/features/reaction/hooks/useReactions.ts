import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

export const useAddReaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ messageId, emoji }: { messageId: string; emoji: string }) => {
      const { error } = await api.POST("/api/messages/{messageId}/reactions", {
        body: { emoji },
        params: { path: { messageId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: (_, { messageId }) => {
      // リアクションとメッセージのクエリを無効化
      void queryClient.invalidateQueries({ queryKey: ["reactions", messageId] });
      void queryClient.invalidateQueries({ queryKey: ["channels"] });
    },
  });
};

export const useRemoveReaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ messageId, emoji }: { messageId: string; emoji: string }) => {
      const { error } = await api.DELETE("/api/messages/{messageId}/reactions/{emoji}", {
        params: { path: { emoji, messageId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: (_, { messageId }) => {
      // リアクションとメッセージのクエリを無効化
      void queryClient.invalidateQueries({ queryKey: ["reactions", messageId] });
      void queryClient.invalidateQueries({ queryKey: ["channels"] });
    },
  });
};
