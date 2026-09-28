import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChartCard } from "./ChartCard";

describe("ChartCard", () => {
  test("グラフと表を切り替えられる", async () => {
    render(
      <ChartCard
        title="日別のメッセージ"
        note="直近 30 日"
        table={{ columns: ["日付", "件数"], rows: [["9/28", "12"]] }}
      >
        <div>グラフ</div>
      </ChartCard>,
    );
    expect(screen.getByRole("heading", { name: "日別のメッセージ" })).toBeInTheDocument();
    expect(screen.getByText("グラフ")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "表で表示" }));
    expect(screen.queryByText("グラフ")).not.toBeInTheDocument();
    expect(screen.getByRole("rowheader", { name: "9/28" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "12" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "グラフで表示" }));
    expect(screen.getByText("グラフ")).toBeInTheDocument();
  });
});
