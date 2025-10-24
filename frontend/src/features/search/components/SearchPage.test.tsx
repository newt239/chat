import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

import { SearchPage } from "./SearchPage";

// モックデータ
const mockChannels = [
  {
    id: "channel-1",
    workspaceId: "workspace-1",
    name: "general",
    description: "一般的な話題のチャンネル",
    isPrivate: false,
    createdBy: "user-1",
    createdAt: "2025-01-01T00:00:00Z",
  },
  {
    id: "channel-2",
    workspaceId: "workspace-1",
    name: "random",
    description: null,
    isPrivate: false,
    createdBy: "user-1",
    createdAt: "2025-01-01T00:00:00Z",
  },
];

const mockMembers = [
  {
    userId: "user-1",
    email: "test@example.com",
    displayName: "テストユーザー",
    avatarUrl: null,
    role: "owner" as const,
    joinedAt: "2025-01-01T00:00:00Z",
  },
];

const mockMessages = [
  {
    id: "message-1",
    channelId: "channel-1",
    userId: "user-1",
    parentId: null,
    body: "こんにちは、テストメッセージです",
    createdAt: "2025-01-01T00:00:00Z",
    editedAt: null,
    deletedAt: null,
  },
];

// モックのカスタムフック
vi.mock("@/features/channel/hooks/useChannel", () => ({
  useChannels: () => ({
    data: mockChannels,
    isLoading: false,
    isError: false,
  }),
}));

vi.mock("@/features/workspace/hooks/useMember", () => ({
  useMembers: () => ({
    data: mockMembers,
    isLoading: false,
    isError: false,
  }),
}));

vi.mock("@/features/message/hooks/useMessage", () => ({
  useMessages: () => ({
    data: { messages: mockMessages, hasMore: false },
    isLoading: false,
    isError: false,
  }),
}));

const createTestRouter = (searchParams: { q?: string; filter?: string }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const rootRoute = createRootRoute();

  const searchRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/app/$workspaceId/search",
    component: SearchPage,
    validateSearch: (search: Record<string, unknown>) => ({
      q: typeof search.q === "string" ? search.q : undefined,
      filter: typeof search.filter === "string" ? search.filter : "all",
    }),
  });

  const router = createRouter({
    routeTree: rootRoute.addChildren([searchRoute]),
    history: createMemoryHistory({
      initialEntries: [
        `/app/workspace-1/search?q=${searchParams.q || ""}&filter=${searchParams.filter || "all"}`,
      ],
    }),
    context: { queryClient },
  });

  return { router, queryClient };
};

describe("SearchPage", () => {
  it("検索クエリがない場合、プレースホルダーメッセージを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "", filter: "all" });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    expect(await screen.findByText("キーワードを入力して検索してください")).toBeInTheDocument();
  });

  it("検索クエリがある場合、検索結果を表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "テスト", filter: "all" });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    expect(await screen.findByText(/「テスト」の検索結果:/)).toBeInTheDocument();
  });

  it("フィルターがmessagesの場合、メッセージのみを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "テスト", filter: "messages" });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    // メッセージセクションが表示される
    expect(await screen.findByText("メッセージ")).toBeInTheDocument();
  });

  it("フィルターがchannelsの場合、チャンネルのみを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "general", filter: "channels" });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    // チャンネルセクションが表示される
    expect(await screen.findByText("チャンネル")).toBeInTheDocument();
    expect(await screen.findByText("general")).toBeInTheDocument();
  });

  it("フィルターがusersの場合、ユーザーのみを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "テストユーザー", filter: "users" });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    // ユーザーセクションが表示される
    expect(await screen.findByText("ユーザー")).toBeInTheDocument();
    expect(await screen.findByText("テストユーザー")).toBeInTheDocument();
  });

  it("検索結果が0件の場合、該当メッセージを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "存在しないキーワード", filter: "all" });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    expect(await screen.findByText("検索結果が見つかりませんでした")).toBeInTheDocument();
  });
});
