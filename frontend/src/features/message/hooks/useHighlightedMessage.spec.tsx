import type { ReactNode } from "react";

import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";
import { z } from "zod";

import { useHighlightedMessage } from "#/features/message/hooks/useHighlightedMessage";

const renderWithChannelRoute = async (url: string) => {
  let hookContent: ReactNode = null;
  const rootRoute = createRootRoute();
  const channelRoute = createRoute({
    component: () => hookContent,
    getParentRoute: () => rootRoute,
    path: "/app/$workspaceId/$channelId",
    validateSearch: z.object({ message: z.string().optional() }),
  });
  const router = createRouter({
    history: createMemoryHistory({ initialEntries: [url] }),
    routeTree: rootRoute.addChildren([channelRoute]),
  });
  await router.load();

  const view = renderHook(() => useHighlightedMessage(true, null), {
    wrapper: ({ children }: { children: ReactNode }) => {
      hookContent = children;
      return <RouterProvider router={router} />;
    },
  });
  await waitFor(() => {
    expect(view.result.current).not.toBeNull();
  });
  return view.result;
};

describe("useHighlightedMessage", () => {
  test("message クエリが無いときは対象を返さない", async () => {
    const result = await renderWithChannelRoute("/app/ws1/ch1");

    expect(result.current.targetMessageId).toBeNull();
    expect(result.current.highlightedId).toBeNull();
  });

  test("message クエリのメッセージをハイライト対象にする", async () => {
    const result = await renderWithChannelRoute("/app/ws1/ch1?message=m1");

    expect(result.current.targetMessageId).toBe("m1");
    expect(result.current.highlightedId).toBe("m1");
  });
});
