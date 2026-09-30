import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import {
  Permission,
  PermissionGrantSchema,
  PermissionService,
} from "#/gen/chat/v1/permission_service_pb";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AdminPermissionsTab } from "./AdminPermissionsTab";

import type { UpdatePermissionRequest } from "#/gen/chat/v1/permission_service_pb";

const grants = [
  create(PermissionGrantSchema, {
    allowed: true,
    permission: Permission.PIN_MESSAGES,
    role: WorkspaceRole.ADMIN,
  }),
  create(PermissionGrantSchema, {
    allowed: true,
    permission: Permission.PIN_MESSAGES,
    role: WorkspaceRole.MEMBER,
  }),
];

const setup = async (myRole: WorkspaceRole) => {
  const update = vi.fn<(req: UpdatePermissionRequest) => void>();
  await renderWithProviders(
    <AdminPermissionsTab workspaceId="ws1" grants={grants} myRole={myRole} />,
    "/app/ws1/admin",
    (routes) => {
      routes.rpc(PermissionService.method.updatePermission, (req) => {
        update(req);
        return {};
      });
    },
  );
  return { update };
};

describe("AdminPermissionsTab", () => {
  test("7 つの操作を並べ、オーナー列は常にオンで変更できない", async () => {
    await setup(WorkspaceRole.ADMIN);
    expect(screen.getAllByRole("rowheader")).toHaveLength(7);
    const owner = screen.getByRole("switch", { name: "オーナー: 他人のメッセージの削除" });
    expect(owner).toBeChecked();
    expect(owner).toBeDisabled();
    expect(screen.getByRole("switch", { name: "メンバー: メッセージのピン留め" })).toBeChecked();
    expect(screen.getByRole("switch", { name: "ゲスト: メッセージのピン留め" })).not.toBeChecked();
  });

  test("管理者は管理者の列を変更できない", async () => {
    await setup(WorkspaceRole.ADMIN);
    expect(screen.getByRole("switch", { name: "管理者: メッセージのピン留め" })).toBeDisabled();
  });

  test("オーナーは管理者の列も変更でき、変更を送る", async () => {
    const { update } = await setup(WorkspaceRole.OWNER);
    const admin = screen.getByRole("switch", { name: "管理者: メッセージのピン留め" });
    expect(admin).toBeEnabled();

    await userEvent.click(screen.getByRole("switch", { name: "メンバー: メッセージのピン留め" }));
    await waitFor(() => {
      expect(update).toHaveBeenCalledWith(
        expect.objectContaining({
          allowed: false,
          permission: Permission.PIN_MESSAGES,
          role: WorkspaceRole.MEMBER,
          workspaceId: "ws1",
        }),
      );
    });
  });
});
