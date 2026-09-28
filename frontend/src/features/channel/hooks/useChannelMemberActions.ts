import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

import type { components } from "#/lib/api/schema";

type ChannelRole = components["schemas"]["ChannelMemberInfo"]["role"];

/** チャンネルメンバーの招待・追放・退出・ロール変更をまとめて提供する */
export const useChannelMemberActions = (channelId: string, workspaceId: string | null) => {
  const queryClient = useQueryClient();

  const invalidate = async () => {
    await queryClient.invalidateQueries({ queryKey: ["channels", channelId, "members"] });
    if (workspaceId !== null) {
      await queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "channels"] });
    }
  };

  const invite = useMutation({
    mutationFn: async ({ userId, role }: { userId: string; role: ChannelRole }) => {
      const { error } = await api.POST("/api/channels/{channelId}/members", {
        body: { role, userId },
        params: { path: { channelId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: async ({ userId }: { userId: string }) => {
      const { error } = await api.DELETE("/api/channels/{channelId}/members/{userId}", {
        params: { path: { channelId, userId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const updateRole = useMutation({
    mutationFn: async ({ userId, role }: { userId: string; role: ChannelRole }) => {
      const { error } = await api.PATCH("/api/channels/{channelId}/members/{userId}/role", {
        body: { role },
        params: { path: { channelId, userId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const join = useMutation({
    mutationFn: async () => {
      const { error } = await api.POST("/api/channels/{channelId}/members/self", {
        params: { path: { channelId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  const leave = useMutation({
    mutationFn: async () => {
      const { error } = await api.DELETE("/api/channels/{channelId}/members/self", {
        params: { path: { channelId } },
      });
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: invalidate,
  });

  return { invite, join, leave, remove, updateRole };
};
