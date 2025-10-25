import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { components } from "@/lib/api/schema";

import { api } from "@/lib/api/client";

export const useChannelMembers = (channelId: string | null) => {
  return useQuery({
    queryKey: ["channels", channelId, "members"],
    queryFn: async (): Promise<components["schemas"]["ChannelMemberInfo"][]> => {
      if (channelId === null) {
        return [];
      }

      const { data, error } = await api.GET("/api/channels/{channelId}/members", {
        params: { path: { channelId } },
      });

      if (error || data === undefined) {
        throw new Error(error?.error ?? "Failed to fetch channel members");
      }

      return data.members;
    },
    enabled: channelId !== null,
  });
};

export const useInviteChannelMember = (channelId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      userId,
      role = "member",
    }: {
      userId: string;
      role?: "member" | "admin";
    }) => {
      const { data, error } = await api.POST("/api/channels/{channelId}/members", {
        params: { path: { channelId } },
        body: { userId, role },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["channels", channelId, "members"] });
    },
  });
};

export const useJoinChannel = (channelId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const { data, error } = await api.POST("/api/channels/{channelId}/members/self", {
        params: { path: { channelId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["channels", channelId, "members"] });
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
};

export const useLeaveChannel = (channelId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const { data, error } = await api.DELETE("/api/channels/{channelId}/members/self", {
        params: { path: { channelId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["channels", channelId, "members"] });
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
};

export const useUpdateChannelMemberRole = (channelId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ userId, role }: { userId: string; role: "member" | "admin" }) => {
      const { data, error } = await api.PATCH(
        "/api/channels/{channelId}/members/{userId}/role",
        {
          params: { path: { channelId, userId } },
          body: { role },
        }
      );

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["channels", channelId, "members"] });
    },
  });
};

export const useRemoveChannelMember = (channelId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ userId }: { userId: string }) => {
      const { data, error } = await api.DELETE("/api/channels/{channelId}/members/{userId}", {
        params: { path: { channelId, userId } },
      });

      if (error) {
        throw new Error(error.error);
      }

      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["channels", channelId, "members"] });
    },
  });
};
