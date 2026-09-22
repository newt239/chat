import type { components } from "#/lib/api/schema";

export const searchFilterValues = ["all", "messages", "channels", "users", "groups"] as const;
export type SearchFilter = (typeof searchFilterValues)[number];

export type WorkspaceSearchResponse = components["schemas"]["WorkspaceSearchResponse"];
