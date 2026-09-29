import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { Switch } from "./Switch";

describe("Switch", () => {
  test("押すとオンになり、状態を通知する", async () => {
    const onChange = vi.fn<(isSelected: boolean) => void>();
    render(<Switch onChange={onChange}>コンパクト表示</Switch>);

    const toggle = screen.getByRole("switch", { name: "コンパクト表示" });
    expect(toggle).not.toBeChecked();

    await userEvent.click(toggle);

    expect(toggle).toBeChecked();
    expect(onChange).toHaveBeenCalledWith(true);
  });

  test("無効なときは切り替えられない", async () => {
    render(<Switch isDisabled>コンパクト表示</Switch>);

    const toggle = screen.getByRole("switch", { name: "コンパクト表示" });
    await userEvent.click(toggle);

    expect(toggle).not.toBeChecked();
  });
});
