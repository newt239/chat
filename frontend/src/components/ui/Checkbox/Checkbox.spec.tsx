import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { Checkbox } from "./Checkbox";

describe("Checkbox", () => {
  test("押すとチェックされ、状態を通知する", async () => {
    const onChange = vi.fn<(isSelected: boolean) => void>();
    render(<Checkbox onChange={onChange}>メンションを通知する</Checkbox>);

    const checkbox = screen.getByRole("checkbox", { name: "メンションを通知する" });
    await userEvent.click(checkbox);

    expect(checkbox).toBeChecked();
    expect(onChange).toHaveBeenCalledWith(true);
  });

  test("一部選択の状態を表せる", () => {
    render(<Checkbox isIndeterminate>一部のチャンネル</Checkbox>);

    expect(screen.getByRole("checkbox", { name: "一部のチャンネル" })).toBePartiallyChecked();
  });
});
