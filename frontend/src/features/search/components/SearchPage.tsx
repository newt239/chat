import { useMemo } from "react";

import { Card, Stack, Text, Loader, SegmentedControl } from "@mantine/core";
import { useNavigate, useParams, useSearch } from "@tanstack/react-router";

import { SearchResultList } from "./SearchResultList";

import { useChannels } from "@/features/channel/hooks/useChannel";
import { useMessages } from "@/features/message/hooks/useMessage";
import { useWorkspaceMembers } from "@/features/workspace/hooks/useWorkspaceMember";

type SearchFilter = "all" | "messages" | "channels" | "users";

export const SearchPage = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId/search" });
  const searchParams = useSearch({ from: "/app/$workspaceId/search" });
  const navigate = useNavigate();

  const query = searchParams.q || "";
  const filter = (searchParams.filter || "all") as SearchFilter;

  const { data: channels, isLoading: isLoadingChannels } = useChannels(workspaceId);
  const { data: members, isLoading: isLoadingMembers } = useWorkspaceMembers(workspaceId);

  // 全チャンネルのメッセージを取得（簡易実装）
  // 本来はサーバー側で検索APIを実装すべき
  const channelIds = useMemo(() => channels?.map((c) => c.id) || [], [channels]);
  const firstChannelId = channelIds[0] || null;
  const { data: messageResponse, isLoading: isLoadingMessages } = useMessages(firstChannelId);

  const isLoading = isLoadingChannels || isLoadingMembers || isLoadingMessages;

  const handleFilterChange = (value: string) => {
    navigate({
      to: "/app/$workspaceId/search",
      params: { workspaceId },
      search: { q: query, filter: value as SearchFilter },
    });
  };

  // 検索結果のフィルタリング
  const filteredResults = useMemo(() => {
    if (!query.trim()) {
      return { messages: [], channels: [], users: [] };
    }

    const lowerQuery = query.toLowerCase();

    const filteredMessages =
      filter === "all" || filter === "messages"
        ? (messageResponse?.messages || []).filter((message) =>
            message.body.toLowerCase().includes(lowerQuery)
          )
        : [];

    const filteredChannels =
      filter === "all" || filter === "channels"
        ? (channels || []).filter(
            (channel) =>
              channel.name.toLowerCase().includes(lowerQuery) ||
              channel.description?.toLowerCase().includes(lowerQuery)
          )
        : [];

    const filteredUsers =
      filter === "all" || filter === "users"
        ? (members || []).filter(
            (member) =>
              member.displayName.toLowerCase().includes(lowerQuery) ||
              member.email.toLowerCase().includes(lowerQuery)
          )
        : [];

    return {
      messages: filteredMessages,
      channels: filteredChannels,
      users: filteredUsers,
    };
  }, [query, filter, messageResponse, channels, members]);

  const totalResults =
    filteredResults.messages.length +
    filteredResults.channels.length +
    filteredResults.users.length;

  return (
    <div className="flex h-full flex-col p-6">
      <Card withBorder padding="lg" radius="md" className="mb-4">
        <Stack gap="md">
          <div>
            <Text size="xl" fw={600}>
              検索結果
            </Text>
            {query && (
              <Text size="sm" c="dimmed" className="mt-1">
                「{query}」の検索結果: {totalResults}件
              </Text>
            )}
          </div>

          <SegmentedControl
            value={filter}
            onChange={handleFilterChange}
            data={[
              { label: "すべて", value: "all" },
              { label: `メッセージ (${filteredResults.messages.length})`, value: "messages" },
              { label: `チャンネル (${filteredResults.channels.length})`, value: "channels" },
              { label: `ユーザー (${filteredResults.users.length})`, value: "users" },
            ]}
          />
        </Stack>
      </Card>

      <div className="flex-1 overflow-y-auto">
        {!query.trim() ? (
          <Card withBorder padding="xl" radius="md" className="flex items-center justify-center">
            <Text c="dimmed" size="sm">
              キーワードを入力して検索してください
            </Text>
          </Card>
        ) : isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader size="sm" />
          </div>
        ) : totalResults === 0 ? (
          <Card withBorder padding="xl" radius="md" className="flex items-center justify-center">
            <Text c="dimmed" size="sm">
              検索結果が見つかりませんでした
            </Text>
          </Card>
        ) : (
          <SearchResultList
            messages={filteredResults.messages}
            channels={filteredResults.channels}
            users={filteredResults.users}
            filter={filter}
            workspaceId={workspaceId}
          />
        )}
      </div>
    </div>
  );
};
