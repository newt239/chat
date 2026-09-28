import type { ReactNode } from "react";

import { createRouterTransport } from "@connectrpc/connect";
import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { render } from "@testing-library/react";
import { Provider as JotaiProvider, createStore } from "jotai";
import { z } from "zod";

import { searchQuerySchema } from "#/features/search/schemas";
import { authAtom } from "#/providers/store/auth";

import type { ConnectRouter } from "@connectrpc/connect";

export const currentUser = {
  avatarUrl: undefined,
  displayName: "Alice",
  email: "alice@example.com",
  id: "00000000-0000-0000-0000-000000000001",
};

// API を routes のハンドラで差し替え、アプリと同じルート構成の中で ui を描画する
export const renderWithProviders = async (
  ui: ReactNode,
  url: string,
  routes: (router: ConnectRouter) => void,
) => {
  const store = createStore();
  store.set(authAtom, { accessToken: "a", refreshToken: "r", user: currentUser });

  const rootRoute = createRootRoute({ component: () => ui });
  const workspaceRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/app/$workspaceId",
  });
  const searchRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/search",
    validateSearch: searchQuerySchema,
  });
  const channelRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/$channelId",
    validateSearch: z.object({ message: z.string().optional() }),
  });
  const router = createRouter({
    history: createMemoryHistory({ initialEntries: [url] }),
    routeTree: rootRoute.addChildren([workspaceRoute.addChildren([searchRoute, channelRoute])]),
  });
  await router.load();

  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <JotaiProvider store={store}>
      <QueryClientProvider client={queryClient}>
        <TransportProvider transport={createRouterTransport(routes)}>
          <RouterProvider router={router} />
        </TransportProvider>
      </QueryClientProvider>
    </JotaiProvider>,
  );
  return { router, store };
};
