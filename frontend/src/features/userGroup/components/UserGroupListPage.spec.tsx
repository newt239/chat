import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserGroupSchema, UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { rightSidePanelViewAtom } from "#/providers/store/ui";
import { renderWithProviders } from "#/test/renderWithProviders";

import { UserGroupListPage } from "./UserGroupListPage";

describe("UserGroupListPage", () => {
  test("グループを押すと右パネルに開き、作成したグループも開く", async () => {
    const { store } = await renderWithProviders(<UserGroupListPage />, "/app/ws1", (routes) => {
      routes.rpc(UserGroupService.method.listUserGroups, () => ({
        userGroups: [create(UserGroupSchema, { id: "g1", name: "frontend" })],
      }));
      routes.rpc(UserGroupService.method.createUserGroup, (req) => ({
        userGroup: create(UserGroupSchema, { id: "g2", name: req.name }),
      }));
      routes.rpc(WorkspaceService.method.getWorkspace, () => ({
        workspace: { id: "ws1", role: WorkspaceRole.ADMIN },
      }));
    });

    await userEvent.click(await screen.findByRole("button", { name: "@frontend" }));
    expect(store.get(rightSidePanelViewAtom)).toEqual({ groupId: "g1", type: "user-group" });

    await userEvent.click(await screen.findByRole("button", { name: "作成" }));
    await userEvent.type(await screen.findByRole("textbox", { name: /グループ名/ }), "design");
    await userEvent.click(screen.getAllByRole("button", { name: "作成" }).at(-1) ?? document.body);
    await waitFor(() => {
      expect(store.get(rightSidePanelViewAtom)).toEqual({ groupId: "g2", type: "user-group" });
    });
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
    expect(screen.queryByRole("button", { name: "作成" })).toBeNull();
  });
});
