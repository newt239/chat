import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Button } from "react-aria-components";
import { describe, expect, test } from "vite-plus/test";

import { Tooltip } from "./Tooltip";

describe("Tooltip", () => {
  test("トリガーにフォーカスすると内容を表示し、Escape で閉じる", async () => {
    render(
      <Tooltip content="最終返信 10:21">
        <Button>スレッド</Button>
      </Tooltip>,
    );
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();

    await userEvent.tab();
    const tooltip = await screen.findByRole("tooltip");
    expect(tooltip).toHaveTextContent("最終返信 10:21");
    expect(screen.getByRole("button", { name: "スレッド" })).toHaveAttribute(
      "aria-describedby",
      tooltip.id,
    );

    await userEvent.keyboard("{Escape}");
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
  });
});
