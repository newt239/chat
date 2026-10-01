import { useState } from "react";

import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { DateTimeField } from "./DateTimeField";

const Controlled = ({ onChange }: { onChange: (value: Date) => void }) => {
  const [value, setValue] = useState(new Date(2026, 8, 30, 9, 0));
  return (
    <DateTimeField
      timeZone="Asia/Tokyo"
      label="送信日時"
      value={value}
      onChange={(next) => {
        setValue(next);
        onChange(next);
      }}
    />
  );
};

describe("DateTimeField", () => {
  test("欄ごとに入力でき、変えた日時を Date で返す", async () => {
    const onChange = vi.fn<(value: Date) => void>();
    render(<Controlled onChange={onChange} />);

    const group = screen.getByRole("group", { name: "送信日時" });
    expect(group).toBeInTheDocument();
    const [first] = screen.getAllByRole("spinbutton");
    if (first === undefined) {
      throw new Error("入力欄がありません");
    }
    // 先頭の欄（en-US では月）を 1 つ進める
    await userEvent.click(first);
    await userEvent.keyboard("{ArrowUp}");

    expect(onChange).toHaveBeenLastCalledWith(new Date(2026, 9, 30, 9, 0));
  });

  test("エラーを表示できる", () => {
    render(
      <DateTimeField
        timeZone="Asia/Tokyo"
        label="送信日時"
        value={new Date(2026, 8, 30, 9, 0)}
        onChange={vi.fn<(value: Date) => void>()}
        errorMessage="現在より後の日時を指定してください"
      />,
    );
    expect(screen.getByText("現在より後の日時を指定してください")).toBeInTheDocument();
  });
});
