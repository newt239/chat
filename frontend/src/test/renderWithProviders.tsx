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

import { adminSearchSchema } from "#/features/admin/schemas";
import { browseChannelsSearchSchema } from "#/features/channel/schemas";
import { workspaceSearchSchema } from "#/features/layout/schemas";
import { jumpDateSchema } from "#/features/message/utils/dateJump";
import { searchQuerySchema } from "#/features/search/schemas";
import { sessionAtom } from "#/providers/store/auth";

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
  store.set(sessionAtom, { accessToken: "a", userId: currentUser.id });

  const rootRoute = createRootRoute({ component: () => ui });
  const workspaceRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/app/$workspaceId",
    validateSearch: workspaceSearchSchema,
  });
  const searchRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/search",
    validateSearch: searchQuerySchema,
  });
  const adminRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/admin",
    validateSearch: adminSearchSchema,
  });
  const insightsRoute = createRoute({ getParentRoute: () => workspaceRoute, path: "/insights" });
  const draftsRoute = createRoute({ getParentRoute: () => workspaceRoute, path: "/drafts" });
  const browseChannelsRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/browse-channels",
    validateSearch: browseChannelsSearchSchema,
  });
  const settingsRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/settings/{-$section}",
  });
  const workspaceSettingsRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/workspace-settings/{-$section}",
  });
  const channelRoute = createRoute({
    getParentRoute: () => workspaceRoute,
    path: "/$channelId",
    validateSearch: z.object({
      date: jumpDateSchema.optional(),
      message: z.string().optional(),
    }),
  });
  const router = createRouter({
    history: createMemoryHistory({ initialEntries: [url] }),
    routeTree: rootRoute.addChildren([
      workspaceRoute.addChildren([
        searchRoute,
        adminRoute,
        insightsRoute,
        draftsRoute,
        browseChannelsRoute,
        settingsRoute,
        workspaceSettingsRoute,
        channelRoute,
      ]),
    ]),
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
