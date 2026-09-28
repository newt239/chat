import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { SearchDateFilter } from "./SearchDateFilter";

type Range = { after: string | null; before: string | null };

describe("SearchDateFilter", () => {
  test("期間を見出しに出し、日付の入力を after / before として渡す", async () => {
    const onChange = vi.fn<(range: Range) => void>();
    render(
      <SearchDateFilter after="2026-09-01" before={null} isInvalid={false} onChange={onChange} />,
    );
    const trigger = screen.getByRole("button", { name: /期間: 2026-09-01〜/ });
    expect(trigger).toHaveAttribute("data-active", "true");

    await userEvent.click(trigger);
    expect(screen.getByLabelText("開始日")).toHaveValue("2026-09-01");
    fireEvent.change(screen.getByLabelText("終了日"), { target: { value: "2026-09-30" } });
    expect(onChange).toHaveBeenLastCalledWith({ after: "2026-09-01", before: "2026-09-30" });

    await userEvent.click(screen.getByRole("button", { name: "期間を解除" }));
    expect(onChange).toHaveBeenLastCalledWith({ after: null, before: null });
  });

  test("未指定なら名前だけを出し、不正な日付があれば警告の見た目にする", () => {
    render(
      <SearchDateFilter
        after={null}
        before={null}
        isInvalid
        onChange={vi.fn<(range: Range) => void>()}
      />,
    );
    const trigger = screen.getByRole("button", { name: "期間" });
    expect(trigger).toHaveAttribute("data-active", "false");
    expect(trigger).toHaveAttribute("data-invalid", "true");
  });
});
