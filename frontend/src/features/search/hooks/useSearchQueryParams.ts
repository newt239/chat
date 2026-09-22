import { useCallback, useMemo } from "react";

import { useSearchParams } from "react-router";
import { z } from "zod";

import { searchFilterValues } from "#/features/search/schemas";

const searchQuerySchema = z.object({
  filter: z.enum(searchFilterValues).catch("all"),
  page: z.coerce.number().int().min(1).catch(1),
  q: z.string().catch(""),
});

type SearchQuery = z.infer<typeof searchQuerySchema>;

export const useSearchQueryParams = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const query = useMemo(
    () =>
      searchQuerySchema.parse({
        filter: searchParams.get("filter") ?? undefined,
        page: searchParams.get("page") ?? undefined,
        q: searchParams.get("q") ?? undefined,
      }),
    [searchParams],
  );

  const updateQuery = useCallback(
    (patch: Partial<SearchQuery>) => {
      const next = { ...query, ...patch };
      setSearchParams({ filter: next.filter, page: String(next.page), q: next.q });
    },
    [query, setSearchParams],
  );

  return { query, updateQuery };
};
