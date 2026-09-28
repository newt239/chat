import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { AlertDialog } from "./AlertDialog";

const renderAlert = () => {
  const onConfirm = vi.fn<() => void>();
  const onOpenChange = vi.fn<(isOpen: boolean) => void>();
  render(
    <AlertDialog
      isOpen
      onOpenChange={onOpenChange}
      title="メッセージを削除しますか？"
      confirmLabel="削除する"
      tone="danger"
      onConfirm={onConfirm}
    >
      元に戻せません。
    </AlertDialog>,
  );
  return { onConfirm, onOpenChange };
};

describe("AlertDialog", () => {
  test("alertdialog として表示し、確定で onConfirm を呼ぶ", async () => {
    const { onConfirm } = renderAlert();

    expect(
      screen.getByRole("alertdialog", { name: "メッセージを削除しますか？" }),
    ).toHaveTextContent("元に戻せません。");
    await userEvent.click(screen.getByRole("button", { name: "削除する" }));

    expect(onConfirm).toHaveBeenCalledOnce();
  });

  test("キャンセルで閉じる操作を通知する", async () => {
    const { onConfirm, onOpenChange } = renderAlert();

    await userEvent.click(screen.getByRole("button", { name: "キャンセル" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
