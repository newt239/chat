import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";

import { ThreadListPage } from "./ThreadListPage";

import { createAppWrapper, createTestQueryClient } from "@/test/utils";

const mockUseParticipatingThreads = vi.fn();

vi.mock("@/features/thread/hooks/useParticipatingThreads", () => ({
  useParticipatingThreads: (params: unknown) => mockUseParticipatingThreads(params),
}));

beforeEach(() => {
  mockUseParticipatingThreads.mockReset();
  mockUseParticipatingThreads.mockReturnValue({
    data: {
      items: [
        {
          thread_id: "t1",
          channel_id: "c1",
          first_message: {
            id: "m1",
            channelId: "c1",
            userId: "u1",
            parentId: null,
            body: "仕様相談です",
            createdAt: "2025-01-01T00:00:00Z",
            editedAt: null,
            deletedAt: null,
            isDeleted: false,
            attachments: [],
            deletedBy: null,
          },
          reply_count: 3,
          last_activity_at: "2025-01-02T00:00:00Z",
          unread_count: 1,
        },
      ],
      next_cursor: undefined,
    },
    isLoading: false,
    isFetching: false,
    error: null,
  });
});

describe("ThreadListPage", () => {
  it("見出しとスレッドカードが表示される", async () => {
    const queryClient = createTestQueryClient();

    const root = createRootRoute();
    const route = createRoute({
      getParentRoute: () => root,
      path: "/app/$workspaceId/threads",
      component: ThreadListPage,
    });
    const router = createRouter({
      routeTree: root.addChildren([route]),
      history: createMemoryHistory({ initialEntries: ["/app/w1/threads"] }),
      context: { queryClient },
    });

    render(<RouterProvider router={router} />, { wrapper: createAppWrapper(queryClient) });

    expect(await screen.findByText("参加中のスレッド")).toBeInTheDocument();
    expect(await screen.findByText("仕様相談です")).toBeInTheDocument();
  });
});
