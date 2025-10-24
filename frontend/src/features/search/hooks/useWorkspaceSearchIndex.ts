import { useQuery } from "@tanstack/react-query";

import type { MessageWithUser } from "@/features/message/types";
import type { components } from "@/lib/api/schema";

import { messagesResponseSchema } from "@/features/message/schemas";
import { apiClient } from "@/lib/api/client";

export interface WorkspaceSearchIndex {
  channels: components["schemas"]["Channel"][];
  members: components["schemas"]["MemberInfo"][];
  messages: MessageWithUser[];
}

const EMPTY_INDEX: WorkspaceSearchIndex = {
  channels: [],
  members: [],
  messages: [],
};

export function useWorkspaceSearchIndex(workspaceId: string | undefined) {
  return useQuery<WorkspaceSearchIndex>({
    queryKey: ["workspace-search-index", workspaceId],
    enabled: typeof workspaceId === "string" && workspaceId.length > 0,
    queryFn: async (): Promise<WorkspaceSearchIndex> => {
      if (!workspaceId) {
        return EMPTY_INDEX;
      }

      const [channelResult, memberResult] = await Promise.all([
        apiClient.GET("/api/workspaces/{id}/channels", {
          params: { path: { id: workspaceId } },
        }),
        apiClient.GET("/api/workspaces/{id}/members", {
          params: { path: { id: workspaceId } },
        }),
      ]);

      if (channelResult.error || channelResult.data === undefined) {
        throw new Error(channelResult.error?.error ?? "Failed to fetch channels");
      }

      if (memberResult.error || memberResult.data === undefined) {
        throw new Error(memberResult.error?.error ?? "Failed to fetch workspace members");
      }

      const channels = Array.isArray(channelResult.data) ? channelResult.data : [];
      const members = Array.isArray(memberResult.data.members)
        ? memberResult.data.members
        : [];

      if (channels.length === 0) {
        return { channels, members, messages: [] };
      }

      const messagesByChannel = await Promise.all(
        channels.map(async (channel) => {
          const { data, error } = await apiClient.GET("/api/channels/{channelId}/messages", {
            params: { path: { channelId: channel.id } },
          });

          if (error || data === undefined) {
            throw new Error(error?.error ?? "Failed to fetch messages for search");
          }

          const parsed = messagesResponseSchema.safeParse(data);

          if (!parsed.success) {
            throw new Error("Unexpected response format when loading search messages");
          }

          return parsed.data.messages;
        })
      );

      const messages: MessageWithUser[] = messagesByChannel.flat();

      return { channels, members, messages };
    },
    staleTime: 60_000,
  });
}
