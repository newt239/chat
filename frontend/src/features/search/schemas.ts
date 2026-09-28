import { z } from "zod";

import { SearchFilter as SearchFilterMessage } from "#/gen/chat/v1/search_service_pb";

export const searchFilterValues = ["all", "messages", "channels", "users", "groups"] as const;
export type SearchFilter = (typeof searchFilterValues)[number];

export const searchFilterMessages: Record<SearchFilter, SearchFilterMessage> = {
  all: SearchFilterMessage.ALL,
  channels: SearchFilterMessage.CHANNELS,
  groups: SearchFilterMessage.GROUPS,
  messages: SearchFilterMessage.MESSAGES,
  users: SearchFilterMessage.USERS,
};

// TanStack Router は search params を JSON としてパースするため page は数値で届く
export const searchQuerySchema = z.object({
  filter: z.enum(searchFilterValues).default("all").catch("all"),
  page: z.number().int().min(1).default(1).catch(1),
  q: z.string().default("").catch(""),
});
