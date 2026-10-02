import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test, vi } from "vite-plus/test";

import { IconToggleButton } from "./IconToggleButton";

test("押した状態を aria-pressed で伝え、押すと反対の値を通知する", async () => {
  const onChange = vi.fn<(isSelected: boolean) => void>();
  render(
    <IconToggleButton label="太字" isSelected onChange={onChange}>
      <svg />
    </IconToggleButton>,
  );

  const button = screen.getByRole("button", { name: "太字" });
  expect(button).toHaveAttribute("aria-pressed", "true");

  await userEvent.click(button);

  expect(onChange).toHaveBeenCalledWith(false);
});
