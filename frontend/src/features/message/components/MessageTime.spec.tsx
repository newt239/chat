import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { MessageTime } from "./MessageTime";

describe("MessageTime", () => {
  test("時刻を表示し、フォーカスすると曜日と秒を含む日時をツールチップで出す", async () => {
    render(<MessageTime date={new Date(2026, 8, 28, 10, 16, 5)} />);
    expect(screen.getByText("10:16")).toBeInTheDocument();
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();

    await userEvent.tab();
    expect(await screen.findByRole("tooltip")).toHaveTextContent("2026年9月28日(月) 10:16:05");
  });
});
