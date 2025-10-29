import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { CreateDMRequest, CreateGroupDMRequest } from "../schemas";
import type { components } from "@/lib/api/schema";

import { api } from "@/lib/api/client";

type DMOutput = components["schemas"]["DMOutput"];

export const useDMs = (workspaceId: string) => {
  return useQuery({
    queryKey: ["dms", workspaceId],
    queryFn: async () => {
      const response = await api.GET("/api/workspaces/{id}/dms", {
        params: {
          path: { id: workspaceId },
        },
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to fetch DMs");
      }

      return response.data as DMOutput[];
    },
    enabled: !!workspaceId,
  });
};

export const useCreateDM = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateDMRequest) => {
      const response = await api.POST("/api/workspaces/{id}/dms", {
        params: {
          path: { id: workspaceId },
        },
        body: data,
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to create DM");
      }

      return response.data as DMOutput;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dms", workspaceId] });
    },
  });
};

export const useCreateGroupDM = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateGroupDMRequest) => {
      const response = await api.POST("/api/workspaces/{id}/group-dms", {
        params: {
          path: { id: workspaceId },
        },
        body: data,
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to create group DM");
      }

      return response.data as DMOutput;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dms", workspaceId] });
    },
  });
};
