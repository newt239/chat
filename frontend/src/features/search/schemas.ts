import { z } from "zod";

import { SearchSort, SearchTarget } from "#/gen/chat/v1/search_service_pb";

export const searchFilterValues = ["all", "messages", "channels", "users", "groups"] as const;
export type SearchFilter = (typeof searchFilterValues)[number];

export const searchFilterMessages: Record<SearchFilter, SearchTarget> = {
  all: SearchTarget.ALL,
  channels: SearchTarget.CHANNELS,
  groups: SearchTarget.GROUPS,
  messages: SearchTarget.MESSAGES,
  users: SearchTarget.USERS,
};

export const searchSortValues = ["newest", "relevance"] as const;
export type SearchSortValue = (typeof searchSortValues)[number];

export const searchSortMessages: Record<SearchSortValue, SearchSort> = {
  newest: SearchSort.NEWEST,
  relevance: SearchSort.RELEVANCE,
};

// q は修飾子を含む入力欄の文字列そのもの。チップはこれを解析して描き、操作したら書き戻す
// TanStack Router は search params を JSON としてパースするため page は数値で届く
export const searchQuerySchema = z.object({
  filter: z.enum(searchFilterValues).default("all").catch("all"),
  page: z.number().int().min(1).default(1).catch(1),
  q: z.string().default("").catch(""),
  // スレッドの返信を含める
  replies: z.boolean().default(true).catch(true),
  sort: z.enum(searchSortValues).default("newest").catch("newest"),
  // in: で指定したチャンネルの下階層を含める
  subs: z.boolean().default(true).catch(true),
});

export type SearchParams = z.infer<typeof searchQuerySchema>;
