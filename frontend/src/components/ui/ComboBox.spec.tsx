import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ComboBox } from "./ComboBox";

const options = [
  { label: "general", value: "general" },
  { label: "frontend", value: "frontend" },
  { label: "backend", value: "backend" },
] as const;

describe("ComboBox", () => {
  test("入力で候補を絞り込み、選んだ値を通知する", async () => {
    const onChange = vi.fn<(value: string | null) => void>();
    render(<ComboBox label="チャンネル" options={options} value={null} onChange={onChange} />);

    await userEvent.type(screen.getByRole("combobox", { name: "チャンネル" }), "end");

    const listbox = await screen.findByRole("listbox");
    expect(listbox).toHaveTextContent("frontend");
    expect(listbox).toHaveTextContent("backend");
    expect(listbox).not.toHaveTextContent("general");

    await userEvent.click(screen.getByRole("option", { name: "backend" }));

    expect(onChange).toHaveBeenCalledWith("backend");
  });

  test("選択中の値を入力欄に表示する", () => {
    render(<ComboBox label="チャンネル" options={options} value="frontend" onChange={() => {}} />);

    expect(screen.getByRole("combobox", { name: "チャンネル" })).toHaveValue("frontend");
  });
});
