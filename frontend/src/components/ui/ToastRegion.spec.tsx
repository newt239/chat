import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { toast } from "./toast";
import { ToastRegion } from "./ToastRegion";

describe("ToastRegion", () => {
  test("toast() の内容を表示し、閉じるボタンで消す", async () => {
    render(<ToastRegion />);

    act(() => {
      toast("保存しました", { description: "設定を更新しました", tone: "success" });
    });

    const alert = await screen.findByRole("alertdialog");
    expect(alert).toHaveTextContent("保存しました");
    expect(alert).toHaveTextContent("設定を更新しました");

    await userEvent.click(screen.getByRole("button", { name: "通知を閉じる" }));

    expect(screen.queryByText("保存しました")).not.toBeInTheDocument();
  });
});
