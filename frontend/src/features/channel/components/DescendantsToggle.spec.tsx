import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { DescendantsToggle } from "./DescendantsToggle";

describe("DescendantsToggle", () => {
  test("押すと集約表示を切り替える", async () => {
    const onChange = vi.fn<(isSelected: boolean) => void>();
    render(<DescendantsToggle count={2} isSelected onChange={onChange} />);

    const toggle = screen.getByRole("button", { name: "下階層を含む" });
    expect(toggle).toHaveAttribute("aria-pressed", "true");
    await userEvent.click(toggle);
    expect(onChange).toHaveBeenCalledWith(false);
  });
});
