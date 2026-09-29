import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ScheduleDialog } from "./ScheduleDialog";

const renderDialog = (initialDate: Date, onConfirm: (scheduledAt: Date) => void) =>
  render(
    <ScheduleDialog
      isOpen
      onOpenChange={vi.fn<(isOpen: boolean) => void>()}
      title="送信日時を指定"
      initialDate={initialDate}
      onConfirm={onConfirm}
      isPending={false}
    >
      <p>本文の欄</p>
    </ScheduleDialog>,
  );

describe("ScheduleDialog", () => {
  test("未来の日時なら確定できる", async () => {
    const onConfirm = vi.fn<(scheduledAt: Date) => void>();
    const future = new Date(Date.now() + 24 * 60 * 60 * 1000);
    future.setSeconds(0, 0);
    renderDialog(future, onConfirm);

    expect(screen.getByText("本文の欄")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "予約する" }));
    expect(onConfirm).toHaveBeenCalledWith(future);
  });

  test("過去の日時ではエラーを出して確定しない", async () => {
    const onConfirm = vi.fn<(scheduledAt: Date) => void>();
    renderDialog(new Date(2020, 0, 1, 9, 0), onConfirm);

    await userEvent.click(screen.getByRole("button", { name: "予約する" }));
    expect(screen.getByText("現在より後の日時を指定してください")).toBeInTheDocument();
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
