import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { MessagesTimelineResponse } from "../types";

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

export const useMessages = (channelId: string | null) =>
  useQuery({
    enabled: channelId !== null,
    queryFn: async (): Promise<MessagesTimelineResponse> => {
      if (channelId === null) {
        return { hasMore: false, messages: [] };
      }

      const { data, error } = await api.GET("/api/channels/{channelId}/messages", {
        params: { path: { channelId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    queryKey: ["channels", channelId, "messages"],
  });

export const useSendMessage = (channelId: string | null) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateMessageInput) => {
      if (channelId === null) {
        throw new Error("チャンネルが選択されていません");
      }

      const { data, error } = await api.POST("/api/channels/{channelId}/messages", {
        body: { attachmentIds: input.attachmentIds, body: input.body },
        params: { path: { channelId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: async () => {
      if (channelId !== null) {
        await queryClient.invalidateQueries({ queryKey: ["channels", channelId, "messages"] });
      }
    },
  });
};

export const useUpdateMessage = (channelId: string | null) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateMessageInput) => {
      const { data, error } = await api.PATCH("/api/messages/{messageId}", {
        body: { body: input.body },
        params: { path: { messageId: input.messageId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: async () => {
      if (channelId !== null) {
        await queryClient.invalidateQueries({ queryKey: ["channels", channelId, "messages"] });
      }
    },
  });
};

export const useDeleteMessage = (channelId: string | null) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: DeleteMessageInput) => {
      const { error } = await api.DELETE("/api/messages/{messageId}", {
        params: { path: { messageId: input.messageId } },
      });

      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: async () => {
      if (channelId !== null) {
        await queryClient.invalidateQueries({ queryKey: ["channels", channelId, "messages"] });
      }
    },
  });
};

export const useUpdateReadState = (channelId: string | null, workspaceId: string | null) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      if (channelId === null) {
        throw new Error("チャンネルが選択されていません");
      }

      const lastReadAt = new Date().toISOString();
      const { error } = await api.POST("/api/channels/{channelId}/reads", {
        body: { lastReadAt },
        params: { path: { channelId } },
      });

      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: async () => {
      // チャンネル一覧を再取得してバッジを更新
      if (workspaceId !== null) {
        await queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "channels"] });
      }
    },
  });
};
