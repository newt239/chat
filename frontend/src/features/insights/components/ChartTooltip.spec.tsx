import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ChartTooltip } from "./ChartTooltip";

describe("ChartTooltip", () => {
  test("指定した位置に値を出し、右端では左にずらす", () => {
    render(<ChartTooltip x={90} y={40} text="9/28 · 12 件" />);
    const tooltip = screen.getByRole("status");
    expect(tooltip).toHaveTextContent("9/28 · 12 件");
    expect(tooltip).toHaveStyle({ left: "90%", top: "40%" });
    expect(tooltip.className).toContain("-translate-x-[calc(100%-8px)]");
  });
});
