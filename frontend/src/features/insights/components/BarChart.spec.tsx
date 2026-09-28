import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { BarChart } from "./BarChart";

const data = [
  { isPartial: false, key: "d1", label: "9/26", tooltip: "9/26 · 4 件", value: 4 },
  { isPartial: false, key: "d2", label: "9/27", tooltip: "9/27 · 7 件", value: 7 },
  { isPartial: true, key: "d3", label: "9/28", tooltip: "9/28 · 2 件（途中）", value: 2 },
];

describe("BarChart", () => {
  test("軸を切りのよい最大値にし、ホバーした棒の値を出す", () => {
    render(<BarChart data={data} height={120} labelEvery={1} ariaLabel="日別のメッセージ" />);
    const chart = screen.getByRole("img", { name: "日別のメッセージ" });
    expect(chart).toHaveTextContent("10");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();

    const target = chart.querySelectorAll("i")[1]?.parentElement;
    if (!target) {
      throw new Error("棒が描画されていない");
    }
    fireEvent.pointerEnter(target);
    expect(screen.getByRole("status")).toHaveTextContent("9/27 · 7 件");
  });
});
