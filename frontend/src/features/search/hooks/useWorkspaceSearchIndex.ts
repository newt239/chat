import { create } from "@bufbuild/protobuf";
import { skipToken, useQuery } from "@connectrpc/connect-query";

import { searchFilterMessages, type SearchFilter } from "#/features/search/schemas";
import {
  ChannelSearchResultSchema,
  MessageSearchResultSchema,
  SearchService,
  UserGroupSearchResultSchema,
  UserSearchResultSchema,
} from "#/gen/chat/v1/search_service_pb";

type WorkspaceSearchParams = {
  workspaceId: string | undefined;
  query: string;
  filter: SearchFilter;
  page: number;
  perPage: number;
};

export const useWorkspaceSearch = ({
  workspaceId,
  query,
  filter,
  page,
  perPage,
}: WorkspaceSearchParams) => {
  const trimmedQuery = query.trim();
  const isEnabled =
    workspaceId !== undefined &&
    workspaceId.length > 0 &&
    trimmedQuery.length > 0 &&
    page > 0 &&
    perPage > 0;

  return useQuery(
    SearchService.method.searchWorkspace,
    isEnabled
      ? { filter: searchFilterMessages[filter], page, perPage, query: trimmedQuery, workspaceId }
      : skipToken,
    {
      retry: 1,
      // 未設定の結果を空として扱い、呼び出し側で存在チェックをしなくて済むようにする
      select: (res) => ({
        channels: res.channels ?? create(ChannelSearchResultSchema),
        groups: res.groups ?? create(UserGroupSearchResultSchema),
        messages: res.messages ?? create(MessageSearchResultSchema),
        users: res.users ?? create(UserSearchResultSchema),
      }),
      staleTime: 30_000,
    },
  );
};
