import { useCallback, useState } from "react";

import type {
  UserGroup,
  UserGroupMember,
  CreateUserGroupInput,
  UpdateUserGroupInput,
} from "../types";

import { api } from "@/lib/api/client";

export const useUserGroups = (workspaceId: string) => {
  const [groups, setGroups] = useState<UserGroup[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchGroups = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.GET("/api/user-groups", {
        params: { query: { workspace_id: workspaceId } },
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to fetch user groups");
      }

      const userGroups = response.data?.userGroups || [];
      setGroups(
        userGroups.map((group) => ({
          ...group,
          description: group.description || undefined,
        }))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch user groups");
    } finally {
      setIsLoading(false);
    }
  }, [workspaceId]);

  const createGroup = useCallback(async (input: CreateUserGroupInput) => {
    try {
      const response = await api.POST("/api/user-groups", {
        body: input,
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to create user group");
      }

      const newGroup = response.data;
      if (newGroup) {
        setGroups((prev) => [
          ...prev,
          {
            ...newGroup,
            description: newGroup.description || undefined,
          },
        ]);
      }

      return newGroup;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create user group");
      throw err;
    }
  }, []);

  const updateGroup = useCallback(async (id: string, input: UpdateUserGroupInput) => {
    try {
      const response = await api.PATCH("/api/user-groups/{id}", {
        params: { path: { id } },
        body: {
          name: input.name || "",
          description: input.description,
        },
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to update user group");
      }

      const updatedGroup = response.data;
      if (updatedGroup) {
        setGroups((prev) =>
          prev.map((group) =>
            group.id === id
              ? {
                  ...updatedGroup,
                  description: updatedGroup.description || undefined,
                }
              : group
          )
        );
      }

      return updatedGroup;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update user group");
      throw err;
    }
  }, []);

  const deleteGroup = useCallback(async (id: string) => {
    try {
      const response = await api.DELETE("/api/user-groups/{id}", {
        params: { path: { id } },
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to delete user group");
      }

      setGroups((prev) => prev.filter((group) => group.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete user group");
      throw err;
    }
  }, []);

  const addMember = useCallback(
    async (groupId: string, input: { email: string; role: "admin" | "member" }) => {
      try {
        const response = await api.POST("/api/user-groups/{id}/members", {
          params: { path: { id: groupId } },
          body: input,
        });

        if (response.error) {
          throw new Error(response.error.error || "Failed to add member");
        }

        return response.data;
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to add member");
        throw err;
      }
    },
    []
  );

  const removeMember = useCallback(async (groupId: string, userId: string) => {
    try {
      const response = await api.DELETE("/api/user-groups/{id}/members", {
        params: {
          path: { id: groupId },
          query: { user_id: userId },
        },
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to remove member");
      }

      return response.data;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to remove member");
      throw err;
    }
  }, []);

  const fetchMembers = useCallback(async (groupId: string): Promise<UserGroupMember[]> => {
    try {
      const response = await api.GET("/api/user-groups/{id}/members", {
        params: { path: { id: groupId } },
      });

      if (response.error) {
        throw new Error(response.error.error || "Failed to fetch group members");
      }

      return response.data?.members || [];
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch group members");
      throw err;
    }
  }, []);

  return {
    groups,
    isLoading,
    error,
    fetchGroups,
    createGroup,
    updateGroup,
    deleteGroup,
    addMember,
    removeMember,
    fetchMembers,
  };
};
