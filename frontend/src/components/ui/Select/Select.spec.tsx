import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { Select } from "./Select";

const options = [
  { label: "ライト", value: "light" },
  { label: "ダーク", value: "dark" },
] as const;

describe("Select", () => {
  test("選んだ選択肢の値を通知する", async () => {
    const onChange = vi.fn<(value: string) => void>();
    render(<Select label="表示モード" options={options} value="light" onChange={onChange} />);

    const trigger = screen.getByRole("button", { name: /表示モード/ });
    expect(trigger).toHaveTextContent("ライト");

    await userEvent.click(trigger);
    await userEvent.click(await screen.findByRole("option", { name: "ダーク" }));

    expect(onChange).toHaveBeenCalledWith("dark");
  });
});
