import { useState } from "react";

import { Card, Stack, Text, Loader, SegmentedControl, Pagination, TextInput } from "@mantine/core";
import { IconSearch } from "@tabler/icons-react";

import { useSearchQueryParams } from "#/features/search/hooks/useSearchQueryParams";
import { useWorkspaceSearch } from "#/features/search/hooks/useWorkspaceSearchIndex";
import { searchFilterValues } from "#/features/search/schemas";
import { useWorkspaceId } from "#/lib/routeParams";

import { SearchResultList } from "./SearchResultList";

const RESULTS_PER_PAGE = 20;

const calculatePages = (total: number, per: number) =>
  Math.max(1, Math.ceil(total / Math.max(1, per)));

export const SearchPage = () => {
  const workspaceId = useWorkspaceId();
  const { query: searchQuery, updateQuery } = useSearchQueryParams();

  const { q: query, filter, page } = searchQuery;
  const trimmedQuery = query.trim();

  const [inputValue, setInputValue] = useState(query);
  const [syncedQuery, setSyncedQuery] = useState(query);

  // URL のクエリが変わったら入力欄に反映する
  if (syncedQuery !== query) {
    setSyncedQuery(query);
    setInputValue(query);
  }

  const handleSearchSubmit = (event: React.SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();
    updateQuery({ page: 1, q: inputValue.trim() });
  };

  const {
    data,
    isLoading: isInitialLoading,
    isFetching,
    error,
  } = useWorkspaceSearch({
    filter,
    page,
    perPage: RESULTS_PER_PAGE,
    query,
    workspaceId,
  });

  const isLoading = isInitialLoading || isFetching;

  const messages = data?.messages.items ?? [];
  const channels = data?.channels.items ?? [];
  const users = data?.users.items ?? [];
  const groups = data?.groups.items ?? [];

  const messageCount = data?.messages.total ?? 0;
  const channelCount = data?.channels.total ?? 0;
  const userCount = data?.users.total ?? 0;
  const groupCount = data?.groups.total ?? 0;

  const countByFilter: Record<typeof filter, number> = {
    all: messageCount + channelCount + userCount + groupCount,
    channels: channelCount,
    groups: groupCount,
    messages: messageCount,
    users: userCount,
  };

  const totalResults = countByFilter[filter];

  const totalPages = (() => {
    if (!data) {
      return 0;
    }

    if (filter === "messages") {
      return calculatePages(messageCount, data.messages.perPage);
    }
    if (filter === "channels") {
      return calculatePages(channelCount, data.channels.perPage);
    }
    if (filter === "users") {
      return calculatePages(userCount, data.users.perPage);
    }
    if (filter === "groups") {
      return calculatePages(groupCount, data.groups.perPage);
    }

    return Math.max(
      calculatePages(messageCount, data.messages.perPage),
      calculatePages(channelCount, data.channels.perPage),
      calculatePages(userCount, data.users.perPage),
      calculatePages(groupCount, data.groups.perPage),
    );
  })();

  const handlePageChange = (value: number) => {
    updateQuery({ page: value });
  };

  const handleFilterChange = (value: string) => {
    const parsed = searchFilterValues.find((filterValue) => filterValue === value);
    if (parsed !== undefined) {
      updateQuery({ filter: parsed, page: 1 });
    }
  };

  const filterLabels: Record<typeof filter, string> = {
    all: "すべて",
    channels: "チャンネル",
    groups: "グループ",
    messages: "メッセージ",
    users: "ユーザー",
  };

  const filterOptions = searchFilterValues.map((value) => ({
    label: `${filterLabels[value]} (${countByFilter[value]})`,
    value,
  }));

  const showPagination =
    trimmedQuery.length > 0 && !isLoading && !error && totalPages > 1 && page <= totalPages;

  return (
    <div className="flex h-full flex-col p-6">
      <Card withBorder padding="lg" radius="md" className="mb-4">
        <Stack gap="md">
          <form onSubmit={handleSearchSubmit}>
            <TextInput
              value={inputValue}
              onChange={(event) => {
                setInputValue(event.currentTarget.value);
              }}
              placeholder="メッセージ、チャンネル、ユーザーを検索"
              leftSection={<IconSearch size={16} />}
              aria-label="検索キーワード"
            />
          </form>

          <div>
            <Text size="xl" fw={600}>
              検索結果
            </Text>
            {trimmedQuery && (
              <Text size="sm" c="dimmed" className="mt-1">
                「{query}」の検索結果: {totalResults}件
              </Text>
            )}
          </div>

          <SegmentedControl value={filter} onChange={handleFilterChange} data={filterOptions} />
        </Stack>
      </Card>

      <div className="flex-1 overflow-y-auto">
        {trimmedQuery ? (
          error ? (
            <Card withBorder padding="xl" radius="md" className="flex items-center justify-center">
              <Text c="red" size="sm">
                検索用データの読み込みに失敗しました
              </Text>
            </Card>
          ) : isLoading ? (
            <div className="flex h-full items-center justify-center">
              <Loader size="sm" />
            </div>
          ) : !data || totalResults === 0 ? (
            <Card withBorder padding="xl" radius="md" className="flex items-center justify-center">
              <Text c="dimmed" size="sm">
                検索結果が見つかりませんでした
              </Text>
            </Card>
          ) : (
            <>
              <SearchResultList
                messages={messages}
                channels={channels}
                users={users}
                groups={groups}
                filter={filter}
                workspaceId={workspaceId}
              />
              {showPagination && (
                <div className="mt-6 flex justify-center">
                  <Pagination
                    value={page}
                    total={totalPages}
                    onChange={handlePageChange}
                    size="sm"
                    aria-label="検索結果ページネーション"
                  />
                </div>
              )}
            </>
          )
        ) : (
          <Card withBorder padding="xl" radius="md" className="flex items-center justify-center">
            <Text c="dimmed" size="sm">
              キーワードを入力して検索してください
            </Text>
          </Card>
        )}
      </div>
    </div>
  );
};
