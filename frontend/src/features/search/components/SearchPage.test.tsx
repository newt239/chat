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

import type { WorkspaceSearchResponse } from "@/features/search/schemas";

import { createAppWrapper, createTestQueryClient } from "@/test/utils";

const mockUseWorkspaceSearch = vi.fn();

vi.mock("@/features/search/hooks/useWorkspaceSearchIndex", () => ({
  useWorkspaceSearch: (params: unknown) => mockUseWorkspaceSearch(params),
  useWorkspaceSearchIndex: (params: unknown) => mockUseWorkspaceSearch(params),
}));

const defaultSearchResponse: WorkspaceSearchResponse = {
  messages: {
    items: [
      {
        id: "message-1",
        channelId: "channel-1",
        userId: "user-1",
        parentId: null,
        body: "こんにちは、テストメッセージです",
        mentions: [],
        groups: [],
        links: [],
        reactions: [],
        attachments: [],
        createdAt: "2025-01-01T00:00:00Z",
        editedAt: null,
        deletedAt: null,
        isDeleted: false,
        deletedBy: null,
        user: {
          id: "user-1",
          displayName: "テストユーザー",
          avatarUrl: null,
        },
      },
    ],
    total: 1,
    page: 1,
    perPage: 20,
    hasMore: false,
  },
  channels: {
    items: [
      {
        id: "channel-1",
        workspaceId: "workspace-1",
        name: "general",
        description: "一般的な話題のチャンネル",
        isPrivate: false,
        createdBy: "user-1",
        createdAt: "2025-01-01T00:00:00Z",
        updatedAt: "2025-01-01T00:00:00Z",
        unreadCount: 0,
        hasMention: false,
      },
    ],
    total: 1,
    page: 1,
    perPage: 20,
    hasMore: false,
  },
  users: {
    items: [
      {
        userId: "user-1",
        email: "test@example.com",
        displayName: "テストユーザー",
        avatarUrl: null,
        role: "owner",
        joinedAt: "2025-01-01T00:00:00Z",
      },
    ],
    total: 1,
    page: 1,
    perPage: 20,
    hasMore: false,
  },
};

const createTestRouter = (searchParams: { q?: string; filter?: string; page?: number }) => {
  const queryClient = createTestQueryClient();

  const rootRoute = createRootRoute();

  const searchRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/app/$workspaceId/search",
    component: SearchPage,
    validateSearch: (search: Record<string, unknown>) => ({
      q: typeof search.q === "string" ? search.q : undefined,
      filter: typeof search.filter === "string" ? search.filter : "all",
      page: typeof search.page === "number" ? search.page : 1,
    }),
  });

  const router = createRouter({
    routeTree: rootRoute.addChildren([searchRoute]),
    history: createMemoryHistory({
      initialEntries: [
        `/app/workspace-1/search?q=${searchParams.q || ""}&filter=${searchParams.filter || "all"}&page=${searchParams.page ?? 1}`,
      ],
    }),
    context: { queryClient },
  });

  return { router, queryClient };
};

beforeEach(() => {
  mockUseWorkspaceSearch.mockReset();
  mockUseWorkspaceSearch.mockReturnValue({
    data: defaultSearchResponse,
    isLoading: false,
    isFetching: false,
    error: null,
  });
});

describe("SearchPage", () => {
  it("検索クエリがない場合、プレースホルダーメッセージを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "", filter: "all" });

    render(<RouterProvider router={router} />, {
      wrapper: createAppWrapper(queryClient),
    });

    expect(await screen.findByText("キーワードを入力して検索してください")).toBeInTheDocument();
  });

  it("検索クエリがある場合、検索結果を表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "テスト", filter: "all" });

    render(<RouterProvider router={router} />, {
      wrapper: createAppWrapper(queryClient),
    });

    expect(await screen.findByText("「テスト」の検索結果: 3件")).toBeInTheDocument();
    expect(screen.getByText("メッセージ")).toBeInTheDocument();
    expect(screen.getByText("チャンネル")).toBeInTheDocument();
    expect(screen.getByText("ユーザー")) .toBeInTheDocument();
  });

  it("フィルターがmessagesの場合、メッセージのみを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "テスト", filter: "messages" });

    render(<RouterProvider router={router} />, {
      wrapper: createAppWrapper(queryClient),
    });

    // メッセージセクションが表示される
    expect(await screen.findByText("メッセージ")).toBeInTheDocument();
    expect(screen.queryByText("チャンネル")).not.toBeInTheDocument();
    expect(screen.queryByText("ユーザー")).not.toBeInTheDocument();
  });

  it("フィルターがchannelsの場合、チャンネルのみを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "general", filter: "channels" });

    render(<RouterProvider router={router} />, {
      wrapper: createAppWrapper(queryClient),
    });

    // チャンネルセクションが表示される
    expect(await screen.findByText("チャンネル")).toBeInTheDocument();
    expect(await screen.findByText("general")).toBeInTheDocument();
    expect(screen.queryByText("メッセージ")).not.toBeInTheDocument();
    expect(screen.queryByText("ユーザー")).not.toBeInTheDocument();
  });

  it("フィルターがusersの場合、ユーザーのみを表示する", async () => {
    const { router, queryClient } = createTestRouter({ q: "テストユーザー", filter: "users" });

    render(<RouterProvider router={router} />, {
      wrapper: createAppWrapper(queryClient),
    });

    // ユーザーセクションが表示される
    expect(await screen.findByText("ユーザー")).toBeInTheDocument();
    expect(await screen.findByText("テストユーザー")).toBeInTheDocument();
    expect(screen.queryByText("メッセージ")).not.toBeInTheDocument();
    expect(screen.queryByText("チャンネル")).not.toBeInTheDocument();
  });

  it("検索結果が0件の場合、該当メッセージを表示する", async () => {
    mockUseWorkspaceSearch.mockReturnValueOnce({
      data: {
        messages: { items: [], total: 0, page: 1, perPage: 20, hasMore: false },
        channels: { items: [], total: 0, page: 1, perPage: 20, hasMore: false },
        users: { items: [], total: 0, page: 1, perPage: 20, hasMore: false },
      },
      isLoading: false,
      isFetching: false,
      error: null,
    });

    const { router, queryClient } = createTestRouter({ q: "存在しないキーワード", filter: "all" });

    render(<RouterProvider router={router} />, {
      wrapper: createAppWrapper(queryClient),
    });

    expect(await screen.findByText("検索結果が見つかりませんでした")).toBeInTheDocument();
  });

  it("複数ページがある場合、ページネーションを表示する", async () => {
    mockUseWorkspaceSearch.mockReturnValueOnce({
      data: {
        messages: { items: defaultSearchResponse.messages.items, total: 45, page: 1, perPage: 20, hasMore: true },
        channels: { items: [], total: 0, page: 1, perPage: 20, hasMore: false },
        users: { items: [], total: 0, page: 1, perPage: 20, hasMore: false },
      },
      isLoading: false,
      isFetching: false,
      error: null,
    });

    const { router, queryClient } = createTestRouter({ q: "テスト", filter: "messages" });

    render(<RouterProvider router={router} />, {
      wrapper: createAppWrapper(queryClient),
    });

    expect(await screen.findByRole("button", { name: "2" })).toBeInTheDocument();
  });
});
