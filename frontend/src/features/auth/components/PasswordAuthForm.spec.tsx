import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { PasswordAuthForm } from "./PasswordAuthForm";

describe("PasswordAuthForm", () => {
  test("送信を通知し、エラーがあれば表示する", async () => {
    const onSubmit = vi.fn<() => void>();
    render(
      <PasswordAuthForm
        onSubmit={onSubmit}
        error={new Error("パスワードが違います")}
        isPending={false}
        submitLabel="ログイン"
      >
        <input aria-label="メール" />
      </PasswordAuthForm>,
    );

    expect(screen.getByRole("textbox", { name: "メール" })).toBeInTheDocument();
    expect(screen.getByText("パスワードが違います")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "ログイン" }));
    expect(onSubmit).toHaveBeenCalledOnce();
  });
});
