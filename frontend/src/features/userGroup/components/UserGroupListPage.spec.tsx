import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserGroupSchema, UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { UserGroupListPage } from "./UserGroupListPage";

describe("UserGroupListPage", () => {
  test("グループを押すと右パネルに開き、作成を押すと作成のダイアログを開く", async () => {
    const { router } = await renderWithProviders(<UserGroupListPage />, "/app/ws1", (routes) => {
      routes.rpc(UserGroupService.method.listUserGroups, () => ({
        userGroups: [create(UserGroupSchema, { id: "g1", name: "frontend" })],
      }));
    });

    await userEvent.click(await screen.findByRole("link", { name: "@frontend" }));
    expect(router.state.location.search).toEqual({ group: "g1" });

    await userEvent.click(screen.getByRole("link", { name: "作成" }));
    expect(router.state.location.search).toEqual({ dialog: "create-group", group: "g1" });
  });
});
