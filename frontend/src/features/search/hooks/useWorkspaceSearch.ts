import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { hasSearchConditions, searchDateRange } from "@chat/search-query";
import { skipToken, useQuery } from "@connectrpc/connect-query";

import { searchFilterMessages, searchSortMessages } from "#/features/search/schemas";
import {
  ChannelSearchResultSchema,
  MessageSearchResultSchema,
  SearchHas,
  SearchService,
  UserGroupSearchResultSchema,
  UserSearchResultSchema,
} from "#/gen/chat/v1/search_service_pb";

import { useResolvedSearchQuery } from "./useResolvedSearchQuery";

import type { SearchParams } from "#/features/search/schemas";

import type { SearchHas as SearchHasValue } from "@chat/search-query";

const hasMessages: Record<SearchHasValue, SearchHas> = {
  file: SearchHas.FILE,
  image: SearchHas.IMAGE,
  link: SearchHas.LINK,
  location: SearchHas.LOCATION,
  video: SearchHas.VIDEO,
};

const toTimestamp = (date: Date | undefined) =>
  date === undefined ? undefined : timestampFromDate(date);

export const useWorkspaceSearch = (workspaceId: string, search: SearchParams, perPage: number) => {
  const resolved = useResolvedSearchQuery(workspaceId, search.q);
  const { query, users, inChannels, isResolving } = resolved;
  const text = query.keywords.join(" ");
  // 一覧を読み込むまでは未解決として扱わない
  const unresolved = isResolving
    ? []
    : [
        ...users.flatMap(({ name, member }) => (member ? [] : [`@${name}`])),
        ...inChannels.flatMap(({ name, channel }) => (channel ? [] : [`#${name}`])),
      ];
  const needsResolve = users.length > 0 || inChannels.length > 0;
  const { after, before } = searchDateRange(query);
  const isEnabled =
    (text.length > 0 || hasSearchConditions(query)) &&
    !(needsResolve && isResolving) &&
    unresolved.length === 0 &&
    query.invalidDates.length === 0;

  const result = useQuery(
    SearchService.method.searchWorkspace,
    isEnabled
      ? {
          messageFilter: {
            after: toTimestamp(after),
            before: toTimestamp(before),
            channelIds: inChannels.flatMap(({ channel }) => (channel ? [channel.id] : [])),
            excludeReplies: !search.replies,
            fromUserIds: users.flatMap(({ member }) => (member ? [member.userId] : [])),
            has: query.has.map((has) => hasMessages[has]),
            includeDescendantChannels: search.subs,
            mentionsMe: query.is.includes("mention"),
            pinnedOnly: query.is.includes("pinned"),
            threadOnly: query.is.includes("thread"),
          },
          page: search.page,
          perPage,
          query: text,
          sort: searchSortMessages[search.sort],
          target: searchFilterMessages[search.filter],
          workspaceId,
        }
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
  return { ...result, isEnabled, resolved, unresolved };
};
