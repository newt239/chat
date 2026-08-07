import { redirect } from "react-router";

import { ResponsiveLayout } from "#/features/layout/components/ResponsiveLayout";
import { SearchPage } from "#/features/search/components/SearchPage";
import { ThreadListPage } from "#/features/thread/components/ThreadListPage";
import { WorkspaceSelection } from "#/features/workspace/components/WorkspaceSelection";
import { paths } from "#/lib/paths";
import { store } from "#/providers/store";
import { isAuthenticatedAtom } from "#/providers/store/auth";
import { ChannelPage } from "#/routes/ChannelPage";
import { LoginPage } from "#/routes/LoginPage";
import { RegisterPage } from "#/routes/RegisterPage";
import { RootLayout } from "#/routes/RootLayout";
import { RouteErrorBoundary } from "#/routes/RouteErrorBoundary";
import { WorkspaceIndexPage } from "#/routes/WorkspaceIndexPage";
import { WorkspaceLayout } from "#/routes/WorkspaceLayout";

import type { RouteObject } from "react-router";

export const routeTree: RouteObject[] = [
  {
    Component: RootLayout,
    ErrorBoundary: RouteErrorBoundary,
    path: "/",
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
        path: "app",
        // /app 配下すべての認証ガード。親の loader が redirect を throw した時点で
        // 子の loader は実行されないため、ガードはこの 1 箇所で足りる。
        loader: () => {
          if (!store.get(isAuthenticatedAtom)) {
            throw redirect(paths.login());
          }
          return null;
        },
        children: [
          { Component: WorkspaceSelection, index: true },
          {
            Component: WorkspaceLayout,
            path: ":workspaceId",
            children: [
              { Component: WorkspaceIndexPage, index: true },
              { Component: SearchPage, path: "search" },
              { Component: ThreadListPage, path: "threads" },
              { Component: ChannelPage, path: ":channelId" },
            ],
          },
        ],
      },
    ],
  },
];
