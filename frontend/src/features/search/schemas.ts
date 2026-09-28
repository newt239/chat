import { z } from "zod";

import type { components } from "#/lib/api/schema";

export const searchFilterValues = ["all", "messages", "channels", "users", "groups"] as const;
export type SearchFilter = (typeof searchFilterValues)[number];

// TanStack Router は search params を JSON としてパースするため page は数値で届く
export const searchQuerySchema = z.object({
  filter: z.enum(searchFilterValues).default("all").catch("all"),
  page: z.number().int().min(1).default(1).catch(1),
  q: z.string().default("").catch(""),
});

export type WorkspaceSearchResponse = components["schemas"]["WorkspaceSearchResponse"];
