import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserGroupSchema, UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { UserGroupNavList } from "./UserGroupNavList";

describe("UserGroupNavList", () => {
  test("グループを並べ、押すと右パネルに詳細を開く", async () => {
    const { router } = await renderWithProviders(
      <UserGroupNavList workspaceId="ws1" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(UserGroupService.method.listUserGroups, () => ({
          userGroups: [create(UserGroupSchema, { id: "g1", name: "frontend" })],
        }));
      },
    );

    await userEvent.click(await screen.findByRole("link", { name: "@frontend" }));
    expect(router.state.location.search).toEqual({ group: "g1" });
  });
});
