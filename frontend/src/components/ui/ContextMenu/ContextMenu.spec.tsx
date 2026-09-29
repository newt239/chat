import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { MenuItem } from "#/components/ui/MenuItem/MenuItem";

import { ContextMenu } from "./ContextMenu";

describe("ContextMenu", () => {
  test("右クリックで開き、項目を選ぶと閉じる", async () => {
    const onCopy = vi.fn<() => void>();
    render(
      <ContextMenu
        aria-label="メッセージの操作"
        menu={<MenuItem onAction={onCopy}>リンクをコピー</MenuItem>}
      >
        <p>メッセージ本文</p>
      </ContextMenu>,
    );
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();

    fireEvent.contextMenu(screen.getByText("メッセージ本文"), { clientX: 40, clientY: 60 });
    await screen.findByRole("menu", { name: "メッセージの操作" });
    await userEvent.click(screen.getByRole("menuitem", { name: "リンクをコピー" }));

    expect(onCopy).toHaveBeenCalledOnce();
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  test("ブラウザ標準のコンテキストメニューを出さない", () => {
    render(
      <ContextMenu aria-label="操作" menu={<MenuItem>コピー</MenuItem>}>
        <p>本文</p>
      </ContextMenu>,
    );

    const notCanceled = fireEvent.contextMenu(screen.getByText("本文"));

    expect(notCanceled).toBe(false);
  });
});
