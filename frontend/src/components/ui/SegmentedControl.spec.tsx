import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test, vi } from "vite-plus/test";

import { SegmentedControl } from "./SegmentedControl";

test("選んだ値を型付きで返し、選択中の項目を示す", async () => {
  const onChange = vi.fn<(value: "light" | "dark") => void>();
  render(
    <SegmentedControl
      label="モード"
      options={[
        { label: "ライト", value: "light" },
        { label: "ダーク", value: "dark" },
      ]}
      value="light"
      onChange={onChange}
    />,
  );

  expect(screen.getByRole("radio", { name: "ライト" })).toBeChecked();
  await userEvent.click(screen.getByRole("radio", { name: "ダーク" }));

  expect(onChange).toHaveBeenCalledWith("dark");
});
