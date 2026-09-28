import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageEditor } from "./MessageEditor";

describe("MessageEditor", () => {
  test("Enter で前後の空白を除いた本文を保存して閉じる", async () => {
    const onSave = vi.fn<(body: string) => Promise<void>>().mockResolvedValue();
    const onClose = vi.fn<() => void>();
    render(<MessageEditor initialBody="before" onSave={onSave} onClose={onClose} />);

    const textbox = screen.getByRole("textbox", { name: "メッセージを編集" });
    await userEvent.clear(textbox);
    await userEvent.type(textbox, " after {Enter}");

    expect(onSave).toHaveBeenCalledWith("after");
    expect(onClose).toHaveBeenCalledOnce();
  });

  test("空にすると保存せずにエラーを出し、Esc で閉じる", async () => {
    const onSave = vi.fn<(body: string) => Promise<void>>();
    const onClose = vi.fn<() => void>();
    render(<MessageEditor initialBody="before" onSave={onSave} onClose={onClose} />);

    const textbox = screen.getByRole("textbox", { name: "メッセージを編集" });
    await userEvent.clear(textbox);
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    expect(screen.getByText("メッセージを入力してください")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();

    await userEvent.type(textbox, "{Escape}");
    expect(onClose).toHaveBeenCalledOnce();
  });
});
