import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton";
import { useWorkspaceSearch } from "#/features/search/hooks/useWorkspaceSearchIndex";

import { SearchResultList } from "./SearchResultList";

import type { SearchFilter } from "#/features/search/schemas";

type SearchResultsPanelProps = {
  workspaceId: string;
  query: string;
  filter: SearchFilter;
};

// 右パネルに出す簡易版の検索結果。1 ページ目だけを表示する
export const SearchResultsPanel = ({ workspaceId, query, filter }: SearchResultsPanelProps) => {
  const { t } = useTranslation();
  const { data, isFetching, error } = useWorkspaceSearch({
    filter,
    page: 1,
    perPage: 10,
    query,
    workspaceId,
  });

  if (query.trim().length === 0) {
    return <p className="m-0 p-4 text-caption text-muted">{t("search.prompt")}</p>;
  }
  if (error) {
    return (
      <p role="alert" className="m-0 p-4 text-caption text-danger">
        {t("search.failed")}
      </p>
    );
  }
  if (isFetching || data === undefined) {
    return (
      <div className="flex flex-col gap-3 p-4">
        <Skeleton className="h-14 w-full rounded-lg" />
        <Skeleton className="h-14 w-full rounded-lg" />
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto bg-surface">
      <SearchResultList
        messages={data.messages.items}
        channels={data.channels.items}
        users={data.users.items}
        groups={data.groups.items}
        filter={filter}
        workspaceId={workspaceId}
      />
    </div>
  );
};
