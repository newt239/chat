import { Card, Stack, Text, Loader, SegmentedControl, Pagination } from "@mantine/core";
import { useNavigate, useParams, useSearch } from "@tanstack/react-router";

import { SearchResultList } from "./SearchResultList";

import { useWorkspaceSearch } from "@/features/search/hooks/useWorkspaceSearchIndex";
import { searchFilterValues, type SearchFilter } from "@/features/search/schemas";

const RESULTS_PER_PAGE = 20;

export const SearchPage = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId/search" });
  const searchParams = useSearch({ from: "/app/$workspaceId/search" });
  const navigate = useNavigate();

  const query = searchParams.q || "";
  const filter = (searchParams.filter || "all") as SearchFilter;
  const pageParam = Number(searchParams.page ?? 1);
  const page = Number.isFinite(pageParam) && pageParam > 0 ? pageParam : 1;
  const trimmedQuery = query.trim();

  const { data, isLoading: isInitialLoading, isFetching, error } = useWorkspaceSearch({
    workspaceId,
    query,
    filter,
    page,
    perPage: RESULTS_PER_PAGE,
  });

  const isLoading = isInitialLoading || isFetching;

  const messages = data?.messages.items ?? [];
  const channels = data?.channels.items ?? [];
  const users = data?.users.items ?? [];

  const messageCount = data?.messages.total ?? 0;
  const channelCount = data?.channels.total ?? 0;
  const userCount = data?.users.total ?? 0;

  const totalResults = filter === "all"
    ? messageCount + channelCount + userCount
    : filter === "messages"
      ? messageCount
      : filter === "channels"
        ? channelCount
        : userCount;

  const totalPages = (() => {
    if (!data) {
      return 0;
    }

    const calculatePages = (total: number, per: number) => Math.max(1, Math.ceil(total / Math.max(1, per)));

    if (filter === "messages") {
      return calculatePages(messageCount, data.messages.perPage);
    }
    if (filter === "channels") {
      return calculatePages(channelCount, data.channels.perPage);
    }
    if (filter === "users") {
      return calculatePages(userCount, data.users.perPage);
    }

    return Math.max(
      calculatePages(messageCount, data.messages.perPage),
      calculatePages(channelCount, data.channels.perPage),
      calculatePages(userCount, data.users.perPage)
    );
  })();

  const handlePageChange = (value: number) => {
    navigate({
      to: "/app/$workspaceId/search",
      params: { workspaceId },
      search: { q: query, filter, page: value },
    });
  };

  const handleFilterChange = (value: string) => {
    navigate({
      to: "/app/$workspaceId/search",
      params: { workspaceId },
      search: { q: query, filter: value as SearchFilter, page: 1 },
    });
  };

  const filterOptions = searchFilterValues.map((value) => {
    const count = value === "messages" ? messageCount : value === "channels" ? channelCount : value === "users" ? userCount : messageCount + channelCount + userCount;
    const label =
      value === "messages"
        ? `メッセージ (${count})`
        : value === "channels"
          ? `チャンネル (${count})`
          : value === "users"
            ? `ユーザー (${count})`
            : `すべて (${count})`;
    return { label, value };
  });

  const showPagination =
    trimmedQuery.length > 0 &&
    !isLoading &&
    !error &&
    totalPages > 1 &&
    page <= totalPages;

  return (
    <div className="flex h-full flex-col p-6">
      <Card withBorder padding="lg" radius="md" className="mb-4">
        <Stack gap="md">
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

          <SegmentedControl
            value={filter}
            onChange={handleFilterChange}
            data={filterOptions}
          />
        </Stack>
      </Card>

      <div className="flex-1 overflow-y-auto">
        {!trimmedQuery ? (
          <Card withBorder padding="xl" radius="md" className="flex items-center justify-center">
            <Text c="dimmed" size="sm">
              キーワードを入力して検索してください
            </Text>
          </Card>
        ) : error ? (
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
              filter={filter}
              workspaceId={workspaceId ?? ""}
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
        )}
      </div>
    </div>
  );
};
