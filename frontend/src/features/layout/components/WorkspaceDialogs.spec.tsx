import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserGroupSchema, UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { WorkspaceDialogs } from "./WorkspaceDialogs";

import type { ConnectRouter } from "@connectrpc/connect";

const groupRoutes = (routes: ConnectRouter) => {
  routes.rpc(UserGroupService.method.listUserGroups, () => ({
    userGroups: [create(UserGroupSchema, { id: "g1", name: "frontend" })],
  }));
  routes.rpc(UserGroupService.method.createUserGroup, (req) => ({
    userGroup: create(UserGroupSchema, { id: "g2", name: req.name }),
  }));
  routes.rpc(WorkspaceService.method.getWorkspace, () => ({
    workspace: { id: "ws1", role: WorkspaceRole.ADMIN },
  }));
};

describe("WorkspaceDialogs", () => {
  test("作成したユーザーグループは右パネルで開く", async () => {
    const { router } = await renderWithProviders(
      <WorkspaceDialogs workspaceId="ws1" />,
      "/app/ws1?dialog=create-group",
      groupRoutes,
    );
    await userEvent.type(await screen.findByRole("textbox", { name: /グループ名/ }), "design");
    await userEvent.click(screen.getByRole("button", { name: "作成" }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({ group: "g2" });
    });
  });

  test("編集するグループは ?group= で決まる", async () => {
    await renderWithProviders(
      <WorkspaceDialogs workspaceId="ws1" />,
      "/app/ws1?dialog=edit-group&group=g1",
      groupRoutes,
    );
    expect(await screen.findByRole("textbox", { name: /グループ名/ })).toHaveValue("frontend");
  });
});
