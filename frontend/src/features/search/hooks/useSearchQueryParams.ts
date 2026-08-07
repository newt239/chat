import { useCallback, useMemo } from "react";

import { useSearchParams } from "react-router";
import { z } from "zod";

import { searchFilterValues } from "#/features/search/schemas";

// 値が欠落・不正な場合は catch でフォールバックする（旧 validateSearch と同等の保証）
const searchQuerySchema = z.object({
  q: z.string().catch(""),
  filter: z.enum(searchFilterValues).catch("all"),
  page: z.coerce.number().int().min(1).catch(1),
});

type SearchQuery = z.infer<typeof searchQuerySchema>;

/** `/app/:workspaceId/search` の URL クエリを検証済みの値として読み書きする。 */
export const useSearchQueryParams = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const query = useMemo(
    () =>
      searchQuerySchema.parse({
        q: searchParams.get("q") ?? undefined,
        filter: searchParams.get("filter") ?? undefined,
        page: searchParams.get("page") ?? undefined,
      }),
    [searchParams],
  );

  const updateQuery = useCallback(
    (patch: Partial<SearchQuery>) => {
      const next = { ...query, ...patch };
      setSearchParams({ q: next.q, filter: next.filter, page: String(next.page) });
    },
    [query, setSearchParams],
  );

  return { query, updateQuery };
};
