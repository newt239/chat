import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { Badge } from "./Badge";

describe("Badge", () => {
  test("既定では件数のバッジとして表示する", () => {
    render(<Badge>12</Badge>);

    expect(screen.getByText("12")).toHaveClass("bg-badge", "text-badge-fg");
  });

  test("tone でラベルの見た目に切り替える", () => {
    render(<Badge tone="tag">BOT</Badge>);

    expect(screen.getByText("BOT")).toHaveClass("bg-sunken", "text-muted");
  });
});
