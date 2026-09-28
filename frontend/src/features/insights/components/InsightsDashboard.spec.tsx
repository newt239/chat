import { create } from "@bufbuild/protobuf";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import {
  GetInsightsResponseSchema,
  InsightService,
  StorageCategory,
} from "#/gen/chat/v1/insight_service_pb";
import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { InsightsDashboard } from "./InsightsDashboard";

import type { GetInsightsRequest } from "#/gen/chat/v1/insight_service_pb";

const insights = create(GetInsightsResponseSchema, {
  activeMembers: { current: 3n, previous: 3n },
  channels: [
    { channelId: "c1", messageCount: 40, name: "general", recentMessageCount: 2 },
    { channelId: "c2", isPrivate: true, messageCount: 10, name: "secret", recentMessageCount: 9 },
  ],
  dailyActivity: [
    { activeMemberCount: 2, date: "2026-09-27", messageCount: 5 },
    { activeMemberCount: 1, date: "2026-09-28", messageCount: 3 },
  ],
  heatmap: [{ hour: 10, messageCount: 8, weekday: 2 }],
  heatmapWeeks: 4,
  memberCount: { current: 4n, previous: 4n },
  messageCount: { current: 50n, previous: 40n },
  myDailyMessages: [{ count: 1, date: "2026-09-28" }],
  storageBreakdown: [
    { bytes: 1024n, category: StorageCategory.IMAGE, fileCount: 2 },
    { bytes: 4096n, category: StorageCategory.VIDEO, fileCount: 1 },
  ],
  storageBytes: { current: 5120n, previous: 0n },
});

const setup = (role: WorkspaceRole) => {
  const request = vi.fn<(req: GetInsightsRequest) => void>();
  return {
    render: () =>
      renderWithProviders(
        <InsightsDashboard workspaceId="ws1" />,
        "/app/ws1/insights",
        (routes) => {
          routes.rpc(InsightService.method.getInsights, (req) => {
            request(req);
            return insights;
          });
          routes.rpc(WorkspaceService.method.getWorkspace, () => ({
            workspace: { id: "ws1", role },
          }));
        },
      ),
    request,
  };
};

describe("InsightsDashboard", () => {
  test("ブラウザのタイムゾーンを送り、各グラフを表示する", async () => {
    const { render, request } = setup(WorkspaceRole.MEMBER);
    await render();

    expect(await screen.findByRole("heading", { name: "日別のメッセージ" })).toBeInTheDocument();
    expect(request).toHaveBeenCalledWith(
      expect.objectContaining({
        timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        workspaceId: "ws1",
      }),
    );
    const popular = screen
      .getByRole("heading", { name: "よく使われているチャンネル" })
      .closest("section");
    if (!popular) {
      throw new Error("カードが見つからない");
    }
    const rows = within(popular).getAllByRole("listitem");
    expect(rows[0]).toHaveTextContent("secret9");

    const storage = screen.getByRole("heading", { name: "ストレージの内訳" }).closest("section");
    if (!storage) {
      throw new Error("カードが見つからない");
    }
    expect(storage).toHaveTextContent("合計 5 KB");
    await userEvent.click(within(storage).getByRole("button", { name: "表で表示" }));
    expect(within(storage).getByRole("rowheader", { name: "動画" })).toBeInTheDocument();

    expect(screen.queryByRole("link", { name: "管理画面を開く" })).not.toBeInTheDocument();
  });

  test("管理者には管理画面へのリンクを出す", async () => {
    const { render } = setup(WorkspaceRole.ADMIN);
    await render();
    expect(await screen.findByRole("link", { name: "管理画面を開く" })).toBeInTheDocument();
  });
});
