import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { Heatmap } from "./Heatmap";

const grid = Array.from({ length: 7 }, (_row, weekday) =>
  Array.from({ length: 24 }, (_cell, hour) => (weekday === 0 && hour === 9 ? 2.5 : 0)),
);

describe("Heatmap", () => {
  test("7 × 24 のセルを描き、ホバーしたセルの平均を出す", () => {
    render(<Heatmap grid={grid} ariaLabel="会話が多い時間帯" />);
    const heatmap = screen.getByRole("img", { name: "会話が多い時間帯" });
    const cells = heatmap.querySelectorAll("i");
    expect(cells).toHaveLength(7 * 24);
    expect(screen.getByText("セルにカーソルを合わせると件数を表示")).toBeInTheDocument();

    fireEvent.pointerEnter(cells.item(9));
    expect(screen.getByText("月 9 時台 · 平均 2.5 件")).toBeInTheDocument();
  });
});
