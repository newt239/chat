import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { TextField } from "./TextField";

describe("TextField", () => {
  test("ラベルと説明が入力欄に関連付けられ、入力を通知する", async () => {
    const onChange = vi.fn<(value: string) => void>();
    render(<TextField label="チャンネル名" description="小文字の英数字" onChange={onChange} />);

    const input = screen.getByRole("textbox", { name: "チャンネル名" });
    expect(input).toHaveAccessibleDescription("小文字の英数字");

    await userEvent.type(input, "dev");

    expect(onChange).toHaveBeenLastCalledWith("dev");
  });

  test("errorMessage を渡すと不正な状態としてメッセージを表示する", () => {
    render(<TextField label="メール" errorMessage="形式が正しくありません" />);

    const input = screen.getByRole("textbox", { name: "メール" });
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByText("形式が正しくありません")).toBeInTheDocument();
  });
});
