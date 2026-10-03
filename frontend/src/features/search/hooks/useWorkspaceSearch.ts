import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { skipToken, useQuery } from "@connectrpc/connect-query";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { isDescendantPath } from "#/features/channel/utils/channelTree";
import { useMembers } from "#/features/member/hooks/useMembers";
import { searchFilterMessages, searchSortMessages } from "#/features/search/schemas";
import {
  hasSearchConditions,
  parseSearchQuery,
  searchDateRange,
} from "#/features/search/utils/searchQuery";
import {
  ChannelSearchResultSchema,
  MessageSearchResultSchema,
  SearchHas,
  SearchService,
  UserGroupSearchResultSchema,
  UserSearchResultSchema,
} from "#/gen/chat/v1/search_service_pb";

import type { SearchParams } from "#/features/search/schemas";
import type { SearchHas as SearchHasValue } from "#/features/search/utils/searchQuery";
import type { WorkspaceMember } from "#/gen/chat/v1/workspace_service_pb";

const hasMessages: Record<SearchHasValue, SearchHas> = {
  file: SearchHas.FILE,
  image: SearchHas.IMAGE,
  link: SearchHas.LINK,
  location: SearchHas.LOCATION,
  video: SearchHas.VIDEO,
};

export const RESULTS_PER_PAGE = 20;

const same = (a: string, b: string | undefined) =>
  b !== undefined && a.toLowerCase() === b.toLowerCase();

const matchesMember = (name: string, member: WorkspaceMember) =>
  same(name, member.displayName) ||
  same(name, member.nickname) ||
  same(name, member.email.split("@")[0]);

const toTimestamp = (date: Date | undefined) =>
  date === undefined ? undefined : timestampFromDate(date);

export const useWorkspaceSearch = (workspaceId: string, search: SearchParams) => {
  const { data: members } = useMembers(workspaceId);
  const { data: channels } = useChannels(workspaceId);
  const query = parseSearchQuery(search.q);
  // 修飾子の名前（from:@名前 / in:#チャンネル名）をメンバー・チャンネルに解決する
  const users = query.from.map((name) => ({
    member: members?.find((member) => matchesMember(name, member)),
    name,
  }));
  const inChannels = query.in.map((name) => ({
    channel: channels?.find((channel) => same(name, channel.name)),
    name,
  }));
  const isResolving = members === undefined || channels === undefined;
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
          perPage: RESULTS_PER_PAGE,
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
  const resolved = {
    channels: channels ?? [],
    hasDescendants: inChannels.some(({ channel }) =>
      channels?.some(
        (other) => channel !== undefined && isDescendantPath(channel.name, other.name),
      ),
    ),
    inChannels,
    members: members ?? [],
    query,
    users,
  };
  return { ...result, isEnabled, resolved, unresolved };
};

export type ResolvedSearchQuery = ReturnType<typeof useWorkspaceSearch>["resolved"];
