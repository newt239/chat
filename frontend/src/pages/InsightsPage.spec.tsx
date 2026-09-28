import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { GetInsightsResponseSchema, InsightService } from "#/gen/chat/v1/insight_service_pb";
import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { InsightsPage } from "./InsightsPage";

describe("InsightsPage", () => {
  test("見出しとインサイトを表示する", async () => {
    await renderWithProviders(<InsightsPage />, "/app/ws1/insights", (routes) => {
      routes.rpc(InsightService.method.getInsights, () => create(GetInsightsResponseSchema));
      routes.rpc(WorkspaceService.method.getWorkspace, () => ({
        workspace: { id: "ws1", role: WorkspaceRole.MEMBER },
      }));
    });
    expect(screen.getByRole("heading", { level: 1, name: "インサイト" })).toBeInTheDocument();
    expect(await screen.findByText("アクティブなメンバー")).toBeInTheDocument();
  });
});
