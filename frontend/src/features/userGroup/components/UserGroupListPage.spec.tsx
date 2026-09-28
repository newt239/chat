import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserGroupSchema, UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { UserGroupListPage } from "./UserGroupListPage";

describe("UserGroupListPage", () => {
  test("グループを押すと右パネルに開き、作成を押すと作成のダイアログを開く", async () => {
    const { router } = await renderWithProviders(<UserGroupListPage />, "/app/ws1", (routes) => {
      routes.rpc(UserGroupService.method.listUserGroups, () => ({
        userGroups: [create(UserGroupSchema, { id: "g1", name: "frontend" })],
      }));
      routes.rpc(WorkspaceService.method.getWorkspace, () => ({
        workspace: { id: "ws1", role: WorkspaceRole.ADMIN },
      }));
    });

    await userEvent.click(await screen.findByRole("link", { name: "@frontend" }));
    expect(router.state.location.search).toEqual({ group: "g1" });

    await userEvent.click(await screen.findByRole("link", { name: "作成" }));
    expect(router.state.location.search).toEqual({ dialog: "create-group", group: "g1" });
  });

  test("管理者でなければ作成ボタンの代わりに案内を出す", async () => {
    await renderWithProviders(<UserGroupListPage />, "/app/ws1", (routes) => {
      routes.rpc(UserGroupService.method.listUserGroups, () => ({ userGroups: [] }));
      routes.rpc(WorkspaceService.method.getWorkspace, () => ({
        workspace: { id: "ws1", role: WorkspaceRole.MEMBER },
      }));
    });
    expect(
      await screen.findAllByText("グループの作成と編集は管理者だけができます"),
    ).not.toHaveLength(0);
    expect(screen.queryByRole("link", { name: "作成" })).toBeNull();
  });
});
