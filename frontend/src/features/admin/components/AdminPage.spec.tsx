import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { AdminMemberSchema, AdminService } from "#/gen/chat/v1/admin_service_pb";
import { PermissionService } from "#/gen/chat/v1/permission_service_pb";
import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AdminPage } from "./AdminPage";

describe("AdminPage", () => {
  test("6 つのタブを URL の tab で切り替える", async () => {
    const { router } = await renderWithProviders(<AdminPage />, "/app/ws1/admin", (routes) => {
      routes.rpc(AdminService.method.listAdminMembers, () => ({
        members: [create(AdminMemberSchema, { displayName: "Bob", userId: "u2" })],
      }));
      routes.rpc(AdminService.method.listAuditLogs, () => ({ logs: [] }));
      routes.rpc(PermissionService.method.getPermissions, () => ({ grants: [] }));
      routes.rpc(WorkspaceService.method.getWorkspace, () => ({
        workspace: { id: "ws1", role: WorkspaceRole.OWNER },
      }));
    });

    expect(screen.getByRole("heading", { level: 1, name: "管理画面" })).toBeInTheDocument();
    expect(screen.getAllByRole("tab").map((tab) => tab.textContent)).toStrictEqual([
      "概要",
      "メンバー",
      "招待",
      "権限",
      "アプリ",
      "監査ログ",
    ]);
    expect(await screen.findByText("投稿の多いメンバー")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("tab", { name: "権限" }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({ tab: "permissions" });
    });
    expect(
      await screen.findByRole("switch", { name: "メンバー: メンバーの招待" }),
    ).toBeInTheDocument();
  });
});
