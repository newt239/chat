import { useCallback, useState } from "react";

import type { MentionSuggestion } from "../types";
import type { components } from "@/lib/api/schema";
import { api } from "@/lib/api/client";

type MemberInfo = components["schemas"]["MemberInfo"];
type UserGroup = components["schemas"]["UserGroup"];

export const useMentionInput = (workspaceId: string) => {
  const [suggestions, setSuggestions] = useState<MentionSuggestion[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [query, setQuery] = useState("");

  const fetchUserSuggestions = useCallback(
    async (searchQuery: string) => {
      if (!searchQuery.trim()) {
        setSuggestions([]);
        return;
      }

      setIsLoading(true);
      try {
        // ワークスペースメンバーを取得（簡略化のため、全メンバーを取得）
        const response = await api.GET("/api/workspaces/{id}/members", {
          params: { path: { id: workspaceId } },
        });

        if (response.error) {
          console.error("Failed to fetch workspace members:", response.error);
          setSuggestions([]);
          return;
        }

        const members = response.data?.members || [];
        const userSuggestions: MentionSuggestion[] = members
          .filter((member: MemberInfo) =>
            member.displayName.toLowerCase().includes(searchQuery.toLowerCase())
          )
          .slice(0, 10) // 最大10件
          .map((member: MemberInfo) => ({
            id: member.userId,
            name: member.displayName,
            type: "user" as const,
            avatarUrl: member.avatarUrl || undefined,
          }));

        setSuggestions(userSuggestions);
      } catch (error) {
        console.error("Failed to fetch user suggestions:", error);
        setSuggestions([]);
      } finally {
        setIsLoading(false);
      }
    },
    [workspaceId]
  );

  const fetchGroupSuggestions = useCallback(
    async (searchQuery: string) => {
      if (!searchQuery.trim()) {
        return;
      }

      try {
        // ユーザーグループを取得
        const response = await api.GET("/api/user-groups", {
          params: { query: { workspace_id: workspaceId } },
        });

        if (response.error) {
          console.error("Failed to fetch user groups:", response.error);
          return;
        }

        const groups = response.data?.userGroups || [];
        const groupSuggestions: MentionSuggestion[] = groups
          .filter((group: UserGroup) =>
            group.name.toLowerCase().includes(searchQuery.toLowerCase())
          )
          .slice(0, 5) // 最大5件
          .map((group: UserGroup) => ({
            id: group.id,
            name: group.name,
            type: "group" as const,
          }));

        setSuggestions((prev) => [...prev, ...groupSuggestions]);
      } catch (error) {
        console.error("Failed to fetch group suggestions:", error);
      }
    },
    [workspaceId]
  );

  const searchMentions = useCallback(
    async (searchQuery: string) => {
      setQuery(searchQuery);

      if (!searchQuery.trim()) {
        setSuggestions([]);
        return;
      }

      setIsLoading(true);
      setSuggestions([]);

      // ユーザーとグループの候補を並行して取得
      await Promise.all([fetchUserSuggestions(searchQuery), fetchGroupSuggestions(searchQuery)]);

      setIsLoading(false);
    },
    [fetchUserSuggestions, fetchGroupSuggestions]
  );

  const clearSuggestions = useCallback(() => {
    setSuggestions([]);
    setQuery("");
  }, []);

  return {
    suggestions,
    isLoading,
    query,
    searchMentions,
    clearSuggestions,
  };
};
