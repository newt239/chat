import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { KpiCard } from "./KpiCard";

describe("KpiCard", () => {
  test("増えるのが良い指標が増えたら良い色で出す", () => {
    render(
      <KpiCard
        label="アクティブなメンバー"
        value="44"
        unit="/ 56 人"
        delta={{ direction: "up", isGoodWhenUp: true, text: "+8% 前期比" }}
      />,
    );
    expect(screen.getByText("44")).toBeInTheDocument();
    expect(screen.getByText("/ 56 人")).toBeInTheDocument();
    expect(screen.getByText("+8% 前期比")).toHaveClass("text-success");
  });

  test("良し悪しを付けない指標は中立の色にし、比較がなければ出さない", () => {
    const { rerender } = render(
      <KpiCard
        label="ストレージ"
        value="1.5 GB"
        unit=""
        delta={{ direction: "up", isGoodWhenUp: null, text: "+10 MB 前期比" }}
      />,
    );
    expect(screen.getByText("+10 MB 前期比")).toHaveClass("text-muted");

    rerender(<KpiCard label="ストレージ" value="1.5 GB" unit="" delta={null} />);
    expect(screen.queryByText("+10 MB 前期比")).not.toBeInTheDocument();
  });
});
