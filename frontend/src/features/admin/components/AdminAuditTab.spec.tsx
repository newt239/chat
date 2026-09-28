import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { downloadText } from "#/features/admin/utils/downloadText";
import {
  AdminMemberSchema,
  AdminService,
  AuditAction,
  AuditLogSchema,
} from "#/gen/chat/v1/admin_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AdminAuditTab } from "./AdminAuditTab";

import type { ExportAuditLogsRequest, ListAuditLogsRequest } from "#/gen/chat/v1/admin_service_pb";

vi.mock("#/features/admin/utils/downloadText", () => ({ downloadText: vi.fn() }));

const bobId = "00000000-0000-0000-0000-000000000002";

const setup = async () => {
  const list = vi.fn<(req: ListAuditLogsRequest) => void>();
  const exportLogs = vi.fn<(req: ExportAuditLogsRequest) => void>();
  await renderWithProviders(
    <AdminAuditTab
      workspaceId="ws1"
      members={[create(AdminMemberSchema, { displayName: "Bob", userId: bobId })]}
    />,
    "/app/ws1/admin",
    (routes) => {
      routes.rpc(AdminService.method.listAuditLogs, (req) => {
        list(req);
        return {
          logs: [
            create(AuditLogSchema, { action: AuditAction.LOGIN, id: "l1", targetLabel: "Bob" }),
          ],
          totalCount: 120,
        };
      });
      routes.rpc(AdminService.method.exportAuditLogs, (req) => {
        exportLogs(req);
        return { content: "a,b", fileName: "audit.csv" };
      });
    },
  );
  return { exportLogs, list };
};

describe("AdminAuditTab", () => {
  test("既定では直近 30 日を 50 件ずつ取得し、ページを送れる", async () => {
    const { list } = await setup();
    expect(await screen.findByText("1–50 / 120 件")).toBeInTheDocument();
    const first = list.mock.calls[0]?.[0];
    expect(first).toMatchObject({ limit: 50, offset: 0, workspaceId: "ws1" });
    expect(first?.since).toBeDefined();

    await userEvent.click(screen.getByRole("button", { name: "次のページ" }));
    await waitFor(() => {
      expect(list).toHaveBeenLastCalledWith(expect.objectContaining({ offset: 50 }));
    });
  });

  test("実行者と操作の種類で絞り込み、同じ条件で CSV を書き出す", async () => {
    const { exportLogs, list } = await setup();
    await screen.findByText("1–50 / 120 件");

    await userEvent.click(screen.getByRole("button", { name: /実行者/ }));
    await userEvent.click(screen.getByRole("option", { name: "Bob" }));
    await userEvent.click(screen.getByRole("button", { name: /操作の種類/ }));
    await userEvent.click(screen.getByRole("option", { name: "権限を変更" }));
    await userEvent.click(screen.getByRole("button", { name: /期間/ }));
    await userEvent.click(screen.getByRole("option", { name: "すべて" }));

    await waitFor(() => {
      expect(list).toHaveBeenLastCalledWith(
        expect.objectContaining({
          actions: [AuditAction.PERMISSION_CHANGED],
          actorId: bobId,
          offset: 0,
        }),
      );
    });
    expect(list.mock.lastCall?.[0].since).toBeUndefined();

    await userEvent.click(screen.getByRole("button", { name: "CSV を書き出す" }));
    await waitFor(() => {
      expect(downloadText).toHaveBeenCalledWith("a,b", "audit.csv", "text/csv;charset=utf-8");
    });
    expect(exportLogs).toHaveBeenCalledWith(
      expect.objectContaining({ actions: [AuditAction.PERMISSION_CHANGED], actorId: bobId }),
    );
  });
});
