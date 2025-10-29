import { useMemo } from "react";

import { useQuery } from "@tanstack/react-query";

import {
	searchFilterValues,
	type SearchFilter,
	workspaceSearchResponseSchema,
	type WorkspaceSearchResponse,
} from "@/features/search/schemas";
import { api } from "@/lib/api/client";

const searchFilterSet = new Set<SearchFilter>(searchFilterValues);

type WorkspaceSearchParams = {
	workspaceId: string | undefined;
	query: string;
	filter: SearchFilter;
	page: number;
	perPage: number;
};

function normalizeFilter(filter: SearchFilter): SearchFilter {
	return searchFilterSet.has(filter) ? filter : "all";
}

export function useWorkspaceSearch(params: WorkspaceSearchParams) {
	const { workspaceId, query, filter, page, perPage } = params;
	const trimmedQuery = useMemo(() => query.trim(), [query]);
	const normalizedFilter = normalizeFilter(filter);

	const isEnabled =
		typeof workspaceId === "string" &&
		workspaceId.length > 0 &&
		trimmedQuery.length > 0 &&
		page > 0 &&
		perPage > 0;

	return useQuery<WorkspaceSearchResponse>({
		queryKey: [
			"workspace-search",
			workspaceId,
			trimmedQuery,
			normalizedFilter,
			page,
			perPage,
		],
		enabled: isEnabled,
		staleTime: 30_000,
		retry: 1,
		queryFn: async () => {
			if (!workspaceId) {
				throw new Error("Workspace ID is required to perform search");
			}

			const { data, error } = await api.GET("/api/workspaces/{workspaceId}/search", {
				params: {
					path: { workspaceId },
					query: {
						q: trimmedQuery,
						filter: normalizedFilter,
						page,
						perPage,
					},
				},
			});

			if (error || data === undefined) {
				throw new Error(error?.error ?? "ワークスペース検索に失敗しました");
			}

			const parsed = workspaceSearchResponseSchema.safeParse(data);
			if (!parsed.success) {
				throw new Error("検索レスポンスの形式が想定と異なります");
			}

			return parsed.data;
		},
	});
}
