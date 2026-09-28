import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen, within } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import {
  AdminMemberSchema,
  AdminService,
  AuditAction,
  AuditLogSchema,
} from "#/gen/chat/v1/admin_service_pb";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AdminOverviewTab } from "./AdminOverviewTab";

const members = [
  create(AdminMemberSchema, {
    displayName: "Alice",
    recentMessageCount: 3,
    role: WorkspaceRole.OWNER,
    userId: "u1",
  }),
  create(AdminMemberSchema, {
    displayName: "Bob",
    recentMessageCount: 9,
    role: WorkspaceRole.MEMBER,
    storageBytes: 4096n,
    userId: "u2",
  }),
  create(AdminMemberSchema, {
    displayName: "Carol",
    role: WorkspaceRole.MEMBER,
    suspendedAt: timestampFromDate(new Date()),
    userId: "u3",
  }),
];

const kpiValue = (label: string) => screen.getByText(label).parentElement;

describe("AdminOverviewTab", () => {
  test("メンバーの集計・内訳・最近の監査ログを表示する", async () => {
    await renderWithProviders(
      <AdminOverviewTab workspaceId="ws1" members={members} />,
      "/app/ws1/admin",
      (routes) => {
        routes.rpc(AdminService.method.listAuditLogs, () => ({
          logs: [create(AuditLogSchema, { action: AuditAction.CHANNEL_CREATED, id: "l1" })],
          totalCount: 1,
        }));
      },
    );
    expect(kpiValue("メンバー")).toHaveTextContent("2人");
    expect(kpiValue("管理者")).toHaveTextContent("1人");
    expect(kpiValue("停止中")).toHaveTextContent("1人");

    const posters = screen.getByRole("heading", { name: "投稿の多いメンバー" }).closest("section");
    if (!posters) {
      throw new Error("カードが見つからない");
    }
    expect(within(posters).getAllByRole("listitem")[0]).toHaveTextContent("Bob9");

    expect(await screen.findByText("チャンネルを作成")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "すべて見る" })).toHaveAttribute(
      "href",
      "/app/ws1/admin?tab=audit",
    );
  });
});
