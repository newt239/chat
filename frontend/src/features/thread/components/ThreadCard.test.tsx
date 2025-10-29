import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

import { ThreadCard } from "./ThreadCard";

import type { ParticipatingThread } from "@/features/thread/schemas";

import { createAppWrapper, createTestQueryClient } from "@/test/utils";

vi.mock("@/lib/api/client", () => ({
  api: { POST: vi.fn().mockResolvedValue({}) },
}));

const sampleThread: ParticipatingThread = {
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
  unread_count: 2,
};

describe("ThreadCard", () => {
  it("基本情報を表示する", async () => {
    const queryClient = createTestQueryClient();

    const root = createRootRoute();
    const route = createRoute({
      getParentRoute: () => root,
      path: "/app/$workspaceId/threads",
      component: () => <ThreadCard thread={sampleThread} />,
    });

    const router = createRouter({
      routeTree: root.addChildren([route]),
      history: createMemoryHistory({ initialEntries: ["/app/w1/threads"] }),
      context: { queryClient },
    });

    render(<RouterProvider router={router} />, { wrapper: createAppWrapper(queryClient) });

    expect(await screen.findByText("仕様相談です")).toBeInTheDocument();
    expect(screen.getByText(/返信/)).toBeInTheDocument();
  });
});
