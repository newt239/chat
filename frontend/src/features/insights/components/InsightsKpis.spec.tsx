import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { GetInsightsResponseSchema } from "#/gen/chat/v1/insight_service_pb";

import { InsightsKpis } from "./InsightsKpis";

describe("InsightsKpis", () => {
  test("KPI を前期と比べて表示する", () => {
    render(
      <InsightsKpis
        insights={create(GetInsightsResponseSchema, {
          activeMembers: { current: 3n, previous: 2n },
          activeRate: { current: 0.75, previous: 0.5 },
          memberCount: { current: 4n, previous: 4n },
          messageCount: { current: 90n, previous: 100n },
          myDailyMessages: [
            { count: 2, date: "2026-09-27" },
            { count: 3, date: "2026-09-28" },
          ],
          storageBytes: { current: 2048n, previous: 1024n },
        })}
      />,
    );
    expect(screen.getByText("/ 4 人")).toBeInTheDocument();
    expect(screen.getByText("+50% 前期比")).toBeInTheDocument();
    expect(screen.getByText("75%")).toBeInTheDocument();
    expect(screen.getByText("+25 pt 前期比")).toBeInTheDocument();
    expect(screen.getByText("-10% 前期比")).toBeInTheDocument();
    expect(screen.getByText("2 KB")).toBeInTheDocument();
    expect(screen.getByText("+1 KB 前期比")).toBeInTheDocument();
    expect(screen.getByText("あなたの投稿（7 日）").parentElement).toHaveTextContent("5件");
  });

  test("前期が 0 のときは比率を出さない", () => {
    render(
      <InsightsKpis
        insights={create(GetInsightsResponseSchema, {
          activeMembers: { current: 1n, previous: 0n },
          messageCount: { current: 5n, previous: 0n },
        })}
      />,
    );
    expect(screen.queryByText(/% 前期比/)).not.toBeInTheDocument();
  });
});
