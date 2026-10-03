import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { AdminMemberSchema } from "#/gen/chat/v1/admin_service_pb";
import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { AdminMembersTab } from "./AdminMembersTab";

import type {
  RemoveMemberRequest,
  UpdateMemberRoleRequest,
} from "#/gen/chat/v1/workspace_service_pb";

const members = [
  create(AdminMemberSchema, {
    displayName: "Alice",
    email: "alice@example.com",
    role: WorkspaceRole.OWNER,
    userId: currentUser.id,
  }),
  create(AdminMemberSchema, {
    displayName: "Bob",
    email: "bob@example.com",
    lastLoginAt: timestampFromDate(new Date()),
    lastLoginIp: "192.0.2.1",
    lastLoginUserAgent:
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36",
    recentMessageCount: 12,
    role: WorkspaceRole.MEMBER,
    storageBytes: 2048n,
    userId: "00000000-0000-0000-0000-000000000002",
  }),
  create(AdminMemberSchema, {
    displayName: "Carol",
    email: "carol@example.com",
    role: WorkspaceRole.GUEST,
    suspendedAt: timestampFromDate(new Date()),
    userId: "00000000-0000-0000-0000-000000000003",
  }),
];

const setup = async () => {
  const updateRole = vi.fn<(req: UpdateMemberRoleRequest) => void>();
  const removeMember = vi.fn<(req: RemoveMemberRequest) => void>();
  await renderWithProviders(
    <AdminMembersTab workspaceId="ws1" members={members} />,
    "/app/ws1/admin",
    (routes) => {
      routes.rpc(WorkspaceService.method.updateMemberRole, (req) => {
        updateRole(req);
        return {};
      });
      routes.rpc(WorkspaceService.method.removeMember, (req) => {
        removeMember(req);
        return {};
      });
    },
  );
  return { removeMember, updateRole };
};

const rowOf = (name: string) => {
  const row = screen.getByText(new RegExp(`^${name}`)).closest("tr");
  if (!row) {
    throw new Error(`${name} の行が見つからない`);
  }
  return row;
};

describe("AdminMembersTab", () => {
  test("ロール・状態・最終ログイン・端末・内訳を表示する", async () => {
    await setup();
    const bob = rowOf("Bob");
    expect(bob).toHaveTextContent("有効");
    expect(bob).toHaveTextContent("192.0.2.1 · Chrome · macOS");
    expect(bob).toHaveTextContent("12");
    expect(bob).toHaveTextContent("2 KB");

    const alice = rowOf("Alice");
    expect(alice).toHaveTextContent("オーナー");
    expect(alice).toHaveTextContent("（あなた）");
    expect(alice).toHaveTextContent("記録なし");
    expect(within(alice).queryByRole("button", { name: "停止" })).not.toBeInTheDocument();

    const carol = rowOf("Carol");
    expect(carol).toHaveTextContent("停止中");
    expect(within(carol).getByRole("button", { name: "再開" })).toBeInTheDocument();
    expect(within(carol).getByRole("button", { name: /Carol のロール/ })).toBeDisabled();
  });

  test("名前とロールで絞り込める", async () => {
    await setup();
    await userEvent.type(screen.getByRole("searchbox", { name: "名前・メールで検索" }), "bob");
    expect(screen.queryByText("Alice")).not.toBeInTheDocument();
    expect(screen.getByText("Bob")).toBeInTheDocument();
    expect(screen.getByText("1 人を表示中", { exact: false })).toBeInTheDocument();
  });

  test("ロールを変更する", async () => {
    const { updateRole } = await setup();
    await userEvent.click(screen.getByRole("button", { name: /Bob のロール/ }));
    await userEvent.click(screen.getByRole("option", { name: "管理者" }));
    await waitFor(() => {
      expect(updateRole).toHaveBeenCalledWith(
        expect.objectContaining({
          role: WorkspaceRole.ADMIN,
          userId: "00000000-0000-0000-0000-000000000002",
          workspaceId: "ws1",
        }),
      );
    });
  });

  test("メンバーを外す前に確認し、確定したときだけ外す", async () => {
    const { removeMember } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "Bob をワークスペースから外す" }));
    const dialog = screen.getByRole("alertdialog", {
      name: "Bob をワークスペースから外しますか？",
    });
    expect(removeMember).not.toHaveBeenCalled();

    await userEvent.click(within(dialog).getByRole("button", { name: "外す" }));
    await waitFor(() => {
      expect(removeMember).toHaveBeenCalledWith(
        expect.objectContaining({
          userId: "00000000-0000-0000-0000-000000000002",
          workspaceId: "ws1",
        }),
      );
    });
  });
});
