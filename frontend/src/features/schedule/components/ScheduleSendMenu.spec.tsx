import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { ScheduleSendMenu } from "./ScheduleSendMenu";

beforeEach(() => {
  vi.useFakeTimers({ now: new Date(2026, 8, 29, 14, 0), shouldAdvanceTime: true });
});

afterEach(() => {
  vi.useRealTimers();
});

describe("ScheduleSendMenu", () => {
  test("プリセットを選ぶとその日時で予約する", async () => {
    const onSchedule = vi.fn<(scheduledAt: Date) => void>();
    render(<ScheduleSendMenu isDisabled={false} onSchedule={onSchedule} />);

    await userEvent.click(screen.getByRole("button", { name: "送信を予約" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "明日の朝 9:00" }));

    expect(onSchedule).toHaveBeenCalledWith(new Date(2026, 8, 30, 9, 0));
  });

  test("日時を指定するダイアログから予約する", async () => {
    const onSchedule = vi.fn<(scheduledAt: Date) => void>();
    render(<ScheduleSendMenu isDisabled={false} onSchedule={onSchedule} />);

    await userEvent.click(screen.getByRole("button", { name: "送信を予約" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "日時を指定…" }));
    await userEvent.click(screen.getByRole("button", { name: "予約する" }));

    expect(onSchedule).toHaveBeenCalledWith(new Date(2026, 8, 30, 9, 0));
  });

  test("送信できないときは開けない", () => {
    render(<ScheduleSendMenu isDisabled onSchedule={vi.fn<(scheduledAt: Date) => void>()} />);
    expect(screen.getByRole("button", { name: "送信を予約" })).toBeDisabled();
  });
});
