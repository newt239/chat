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

/** アプリ内の遷移先を組み立てる。 React Router の `to` は単なる文字列で型検査が効かないため、 遷移先の組み立てをここに集約して存在しないパスへのリンクを防ぐ。 */
export const paths = {
  login: () => "/login",
  register: () => "/register",
  app: () => "/app",
  workspace: (workspaceId: string) => `/app/${workspaceId}`,
  channel: (workspaceId: string, channelId: string, messageId?: string) =>
    withQuery(`/app/${workspaceId}/${channelId}`, { message: messageId }),
  threads: (workspaceId: string) => `/app/${workspaceId}/threads`,
  search: (workspaceId: string, query: SearchQuery = {}) =>
    withQuery(`/app/${workspaceId}/search`, {
      q: query.q,
      filter: query.filter,
      page: query.page === undefined ? undefined : String(query.page),
    }),
};
