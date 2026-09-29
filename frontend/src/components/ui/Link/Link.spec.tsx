import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { LinkButton } from "#/components/ui/LinkButton/LinkButton";

import { Link } from "./Link";

const renderWithRouter = async () => {
  const rootRoute = createRootRoute();
  const indexRoute = createRoute({
    component: () => (
      <>
        <Link
          to="/app/$workspaceId/$channelId"
          params={{ channelId: "general", workspaceId: "ws" }}
        >
          general
        </Link>
        <LinkButton
          to="/app/$workspaceId/$channelId"
          params={{ channelId: "random", workspaceId: "ws" }}
          variant="secondary"
        >
          random
        </LinkButton>
      </>
    ),
    getParentRoute: () => rootRoute,
    path: "/",
  });
  const channelRoute = createRoute({
    component: () => <p>channel page</p>,
    getParentRoute: () => rootRoute,
    path: "/app/$workspaceId/$channelId",
  });
  const router = createRouter({
    history: createMemoryHistory({ initialEntries: ["/"] }),
    routeTree: rootRoute.addChildren([indexRoute, channelRoute]),
  });
  await router.load();
  render(<RouterProvider router={router} />);
  return router;
};

describe("Link / LinkButton", () => {
  test("パラメータから組み立てた href を持ち、押すと遷移する", async () => {
    const router = await renderWithRouter();

    const link = await screen.findByRole("link", { name: "general" });
    expect(link).toHaveAttribute("href", "/app/ws/general");
    await userEvent.click(link);

    expect(await screen.findByText("channel page")).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/app/ws/general");
  });

  test("LinkButton はボタンの見た目のリンクとして描画する", async () => {
    await renderWithRouter();

    const link = await screen.findByRole("link", { name: "random" });
    expect(link).toHaveAttribute("href", "/app/ws/random");
    expect(link).toHaveClass("bg-surface");
  });
});
