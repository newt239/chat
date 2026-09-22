import type { SearchFilter } from "#/features/search/schemas";

type SearchQuery = {
  q?: string;
  filter?: SearchFilter;
  page?: number;
};

const withQuery = (pathname: string, query: Record<string, string | undefined>) => {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== "") {
      params.set(key, value);
    }
  }
  const queryString = params.toString();
  return queryString === "" ? pathname : `${pathname}?${queryString}`;
};

export const paths = {
  app: () => "/app",
  channel: (workspaceId: string, channelId: string, messageId?: string) =>
    withQuery(`/app/${workspaceId}/${channelId}`, { message: messageId }),
  login: () => "/login",
  register: () => "/register",
  search: (workspaceId: string, query: SearchQuery = {}) =>
    withQuery(`/app/${workspaceId}/search`, {
      filter: query.filter,
      page: query.page === undefined ? undefined : String(query.page),
      q: query.q,
    }),
  threads: (workspaceId: string) => `/app/${workspaceId}/threads`,
  workspace: (workspaceId: string) => `/app/${workspaceId}`,
};
