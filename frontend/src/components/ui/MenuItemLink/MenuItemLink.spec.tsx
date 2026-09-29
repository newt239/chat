import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test } from "vite-plus/test";

import { Button } from "#/components/ui/Button/Button";
import { Menu } from "#/components/ui/Menu/Menu";

import { MenuItemLink } from "./MenuItemLink";

test("パラメータから組み立てた href を持つ、新しいタブで開くメニュー項目になる", async () => {
  const rootRoute = createRootRoute();
  const indexRoute = createRoute({
    component: () => (
      <Menu trigger={<Button>その他</Button>}>
        <MenuItemLink
          to="/app/$workspaceId/$channelId"
          params={{ channelId: "general", workspaceId: "ws" }}
          target="_blank"
        >
          新しいタブで開く
        </MenuItemLink>
      </Menu>
    ),
    getParentRoute: () => rootRoute,
    path: "/",
  });
  const router = createRouter({
    history: createMemoryHistory({ initialEntries: ["/"] }),
    routeTree: rootRoute.addChildren([indexRoute]),
  });
  await router.load();
  render(<RouterProvider router={router} />);

  await userEvent.click(await screen.findByRole("button", { name: "その他" }));
  const item = screen.getByRole("menuitem", { name: "新しいタブで開く" });

  expect(item).toHaveAttribute("href", "/app/ws/general");
  expect(item).toHaveAttribute("target", "_blank");
});
