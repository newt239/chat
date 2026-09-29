import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { Calendar } from "./Calendar";

const now = new Date();
const firstOfMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-01`;
// セルの名前は「Tuesday, September 1, 2026」のような曜日付きの日付になる
const dayButton = (day: number) =>
  screen.getByRole("button", {
    name: new RegExp(
      new Intl.DateTimeFormat("en-US", { day: "numeric", month: "long", year: "numeric" }).format(
        new Date(now.getFullYear(), now.getMonth(), day),
      ),
    ),
  });

describe("Calendar", () => {
  test("選んだ日を YYYY-MM-DD で返す", async () => {
    const onChange = vi.fn<(date: string) => void>();
    render(<Calendar aria-label="日付" maxDate="9999-12-31" onChange={onChange} />);

    await userEvent.click(dayButton(1));

    expect(onChange).toHaveBeenCalledWith(firstOfMonth);
  });

  test("maxDate より後の日は選べない", async () => {
    const onChange = vi.fn<(date: string) => void>();
    render(<Calendar aria-label="日付" maxDate={firstOfMonth} onChange={onChange} />);

    await userEvent.click(dayButton(2));

    expect(onChange).not.toHaveBeenCalled();
  });
});
