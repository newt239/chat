import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { AuditAction, AuditLogSchema } from "#/gen/chat/v1/admin_service_pb";

import { AuditLogTable } from "./AuditLogTable";

const createdAt = timestampFromDate(new Date(2026, 8, 28, 10, 16));

describe("AuditLogTable", () => {
  test("操作の種類と詳細を表示名に置き換える", () => {
    render(
      <AuditLogTable
        logs={[
          create(AuditLogSchema, {
            action: AuditAction.PERMISSION_CHANGED,
            actor: { displayName: "Alice", id: "u1" },
            createdAt,
            id: "l1",
            ipAddress: "192.0.2.1",
            metadata: { allowed: "false", permission: "pin_messages" },
            targetLabel: "member",
            targetType: "role",
          }),
          create(AuditLogSchema, {
            action: AuditAction.MEMBER_ROLE_CHANGED,
            actor: { displayName: "Alice", id: "u1" },
            createdAt,
            id: "l2",
            metadata: { from: "member", to: "admin" },
            targetLabel: "Bob",
            targetType: "user",
          }),
          create(AuditLogSchema, {
            action: AuditAction.LOGIN_FAILED,
            createdAt,
            id: "l3",
            targetLabel: "Bob",
            targetType: "user",
          }),
          create(AuditLogSchema, {
            action: AuditAction.DATA_EXPORTED,
            actor: { displayName: "Alice", id: "u1" },
            createdAt,
            id: "l4",
            metadata: { count: "12", format: "csv", kind: "audit_log" },
            targetLabel: "監査ログ",
            targetType: "data",
          }),
        ]}
      />,
    );
    const rows = screen.getAllByRole("row").slice(1);
    expect(rows[0]).toHaveTextContent("2026年9月28日 10:16");
    expect(rows[0]).toHaveTextContent("権限を変更");
    expect(rows[0]).toHaveTextContent("メンバー");
    expect(rows[0]).toHaveTextContent("メッセージのピン留め: 不可");
    expect(rows[0]).toHaveTextContent("192.0.2.1 · 不明な端末");
    expect(rows[1]).toHaveTextContent("メンバー → 管理者");
    expect(rows[2]).toHaveTextContent("不明");
    expect(rows[2]).toHaveTextContent("ログインに失敗");
    expect(rows[3]).toHaveTextContent("監査ログ");
    expect(rows[3]).toHaveTextContent("12 件");
  });

  test("ログがなければ案内を出す", () => {
    render(<AuditLogTable logs={[]} />);
    expect(screen.getByText("条件に合う監査ログはありません")).toBeInTheDocument();
  });
});
