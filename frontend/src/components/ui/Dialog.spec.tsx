import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { Button } from "./Button";
import { Dialog } from "./Dialog";

describe("Dialog", () => {
  test("開いているときだけタイトルと内容を表示する", () => {
    const { rerender } = render(
      <Dialog isOpen={false} onOpenChange={() => {}} title="チャンネルを作成">
        本文
      </Dialog>,
    );
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

    rerender(
      <Dialog
        isOpen
        onOpenChange={() => {}}
        title="チャンネルを作成"
        footer={<Button>作成</Button>}
      >
        本文
      </Dialog>,
    );

    const dialog = screen.getByRole("dialog", { name: "チャンネルを作成" });
    expect(dialog).toHaveTextContent("本文");
    expect(screen.getByRole("button", { name: "作成" })).toBeInTheDocument();
  });

  test("閉じるボタンと Escape で閉じる操作を通知する", async () => {
    const onOpenChange = vi.fn<(isOpen: boolean) => void>();
    render(
      <Dialog isOpen onOpenChange={onOpenChange} title="設定">
        本文
      </Dialog>,
    );

    await userEvent.click(screen.getByRole("button", { name: "閉じる" }));
    expect(onOpenChange).toHaveBeenLastCalledWith(false);

    onOpenChange.mockClear();
    await userEvent.keyboard("{Escape}");
    expect(onOpenChange).toHaveBeenLastCalledWith(false);
  });
});
