import { redirect } from "react-router";

import { ResponsiveLayout } from "#/features/layout/components/ResponsiveLayout";
import { SearchPage } from "#/features/search/components/SearchPage";
import { ThreadListPage } from "#/features/thread/components/ThreadListPage";
import { WorkspaceSelection } from "#/features/workspace/components/WorkspaceSelection";
import { paths } from "#/lib/paths";
import { ChannelPage } from "#/pages/ChannelPage";
import { LoginPage } from "#/pages/LoginPage";
import { NotFoundPage } from "#/pages/NotFoundPage";
import { RegisterPage } from "#/pages/RegisterPage";
import { RootLayout } from "#/pages/RootLayout";
import { RouteErrorBoundary } from "#/pages/RouteErrorBoundary";
import { WorkspaceIndexPage } from "#/pages/WorkspaceIndexPage";
import { WorkspaceLayout } from "#/pages/WorkspaceLayout";
import { store } from "#/providers/store";
import { isAuthenticatedAtom } from "#/providers/store/auth";

import type { RouteObject } from "react-router";

export const routeTree: RouteObject[] = [
  {
    Component: RootLayout,
    ErrorBoundary: RouteErrorBoundary,
    children: [
      {
        index: true,
        loader: () => {
          throw redirect(store.get(isAuthenticatedAtom) ? paths.app() : paths.login());
        },
      },
      { Component: LoginPage, path: "login" },
      { Component: RegisterPage, path: "register" },
      {
        Component: ResponsiveLayout,
        children: [
          { Component: WorkspaceSelection, index: true },
          {
            Component: WorkspaceLayout,
            children: [
              { Component: WorkspaceIndexPage, index: true },
              { Component: SearchPage, path: "search" },
              { Component: ThreadListPage, path: "threads" },
              { Component: ChannelPage, path: ":channelId" },
            ],
            path: ":workspaceId",
          },
        ],
        // 親が redirect を throw すると子の loader は実行されないため、ガードはここだけで足りる
        loader: () => {
          if (!store.get(isAuthenticatedAtom)) {
            throw redirect(paths.login());
          }
          return null;
        },
        path: "app",
      },
      { Component: NotFoundPage, path: "*" },
    ],
    path: "/",
  },
];
