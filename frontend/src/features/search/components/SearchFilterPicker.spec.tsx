import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { SearchFilterPicker } from "./SearchFilterPicker";

const options = [
  { label: "Alice", value: "u1" },
  { label: "Bob", value: "u2" },
] as const;

describe("SearchFilterPicker", () => {
  test("選択中の値を見出しに出し、選んだ値を渡す", async () => {
    const onToggle = vi.fn<(value: string) => void>();
    render(
      <SearchFilterPicker
        label="投稿者"
        summary="Alice"
        options={options}
        selected={["u1"]}
        onToggle={onToggle}
        isSearchable
        isInvalid={false}
      />,
    );
    const trigger = screen.getByRole("button", { name: "投稿者: Alice" });
    expect(trigger).toHaveAttribute("data-active", "true");

    await userEvent.click(trigger);
    expect(screen.getByRole("menuitemcheckbox", { name: "Alice" })).toHaveAttribute(
      "aria-checked",
      "true",
    );
    await userEvent.type(screen.getByRole("searchbox", { name: "絞り込む" }), "bo");
    expect(screen.queryByRole("menuitemcheckbox", { name: "Alice" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("menuitemcheckbox", { name: "Bob" }));
    expect(onToggle).toHaveBeenCalledWith("u2");
  });

  test("未選択なら名前だけを出し、解決できなければ警告の見た目にする", () => {
    render(
      <SearchFilterPicker
        label="チャンネル"
        summary={null}
        options={options}
        selected={[]}
        onToggle={vi.fn<(value: string) => void>()}
        isSearchable={false}
        isInvalid
      />,
    );
    const trigger = screen.getByRole("button", { name: "チャンネル" });
    expect(trigger).toHaveAttribute("data-active", "false");
    expect(trigger).toHaveAttribute("data-invalid", "true");
  });
});
