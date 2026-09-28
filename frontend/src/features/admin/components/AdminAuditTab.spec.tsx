import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
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

const setup = async (url = "/app/ws1/admin?tab=audit") => {
  const list = vi.fn<(req: ListAuditLogsRequest) => void>();
  const exportLogs = vi.fn<(req: ExportAuditLogsRequest) => void>();
  const { router } = await renderWithProviders(
    <AdminAuditTab
      workspaceId="ws1"
      members={[create(AdminMemberSchema, { displayName: "Bob", userId: bobId })]}
    />,
    url,
    (routes) => {
      routes.rpc(AdminService.method.listAuditLogs, (req) => {
        list(req);
        // 1 ページ目だけ続きがある
        return req.pageToken === ""
          ? {
              logs: [create(AuditLogSchema, { action: AuditAction.LOGIN, id: "l1" })],
              nextPageToken: "next",
            }
          : { logs: [create(AuditLogSchema, { action: AuditAction.LOGIN, id: "l2" })] };
      });
      routes.rpc(AdminService.method.exportAuditLogs, (req) => {
        exportLogs(req);
        return { content: "a,b", fileName: "audit.csv" };
      });
    },
  );
  return { exportLogs, list, router };
};

describe("AdminAuditTab", () => {
  test("既定では直近 30 日を 50 件ずつ取得し、ページトークンで続きを読み込む", async () => {
    const { list } = await setup();
    expect(await screen.findByText("1 件を表示中")).toBeInTheDocument();
    const first = list.mock.calls[0]?.[0];
    expect(first).toMatchObject({ limit: 50, pageToken: "", workspaceId: "ws1" });
    expect(first?.since).toBeDefined();

    await userEvent.click(screen.getByRole("button", { name: "さらに読み込む" }));
    expect(await screen.findByText("2 件を表示中")).toBeInTheDocument();
    expect(list).toHaveBeenLastCalledWith(expect.objectContaining({ pageToken: "next" }));
    expect(screen.queryByRole("button", { name: "さらに読み込む" })).toBeNull();
  });

  test("絞り込みを URL に保持し、同じ条件で CSV を書き出す", async () => {
    const { exportLogs, list, router } = await setup();
    await screen.findByText("1 件を表示中");

    await userEvent.click(screen.getByRole("button", { name: /実行者/ }));
    await userEvent.click(screen.getByRole("option", { name: "Bob" }));
    await userEvent.click(screen.getByRole("button", { name: /操作の種類/ }));
    await userEvent.click(screen.getByRole("option", { name: "権限を変更" }));
    await userEvent.click(screen.getByRole("button", { name: /期間/ }));
    await userEvent.click(screen.getByRole("option", { name: "すべて" }));

    await waitFor(() => {
      expect(list).toHaveBeenLastCalledWith(
        expect.objectContaining({ actions: [AuditAction.PERMISSION_CHANGED], actorId: bobId }),
      );
    });
    expect(list.mock.lastCall?.[0].since).toBeUndefined();
    expect(router.state.location.search).toMatchObject({
      action: "permissionChanged",
      actor: bobId,
      period: "all",
      tab: "audit",
    });

    await userEvent.click(screen.getByRole("button", { name: "CSV を書き出す" }));
    await waitFor(() => {
      expect(downloadText).toHaveBeenCalledWith("a,b", "audit.csv", "text/csv;charset=utf-8");
    });
    expect(exportLogs).toHaveBeenCalledWith(
      expect.objectContaining({ actions: [AuditAction.PERMISSION_CHANGED], actorId: bobId }),
    );
  });

  test("日時を指定した期間で取得し、開始が終了より後なら知らせる", async () => {
    const { list } = await setup(
      "/app/ws1/admin?tab=audit&period=custom&since=2026-09-01T09:00&until=2026-09-02T18:30",
    );
    await screen.findByText("1 件を表示中");
    const req = list.mock.calls[0]?.[0];
    expect(req?.since && timestampDate(req.since)).toEqual(new Date("2026-09-01T09:00"));
    expect(req?.until && timestampDate(req.until)).toEqual(new Date("2026-09-02T18:30"));
    expect(screen.getByLabelText("開始")).toHaveValue("2026-09-01T09:00");
    expect(screen.queryByText("開始は終了より前にしてください")).toBeNull();

    const until = screen.getByLabelText("終了");
    await userEvent.clear(until);
    await userEvent.type(until, "2026-08-01T00:00");
    expect(await screen.findByText("開始は終了より前にしてください")).toBeInTheDocument();
  });
});
