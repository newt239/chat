import { useState } from "react";

import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChannelNameField } from "./ChannelNameField";

const Field = ({ errorMessage }: { errorMessage: string | null }) => {
  const [value, setValue] = useState("");
  return (
    <ChannelNameField
      label="名前"
      prefix="#dev/"
      value={value}
      onChange={setValue}
      description="説明文"
      errorMessage={errorMessage}
    />
  );
};

describe("ChannelNameField", () => {
  test("接頭辞を表示し、入力を小文字にそろえる", async () => {
    render(<Field errorMessage={null} />);
    expect(screen.getByText("#dev/")).toBeInTheDocument();
    expect(screen.getByText("説明文")).toBeInTheDocument();

    const input = screen.getByRole("textbox", { name: "名前" });
    await userEvent.type(input, "Web");
    expect(input).toHaveValue("web");
  });

  test("エラーがあれば説明の代わりに表示し、不正な状態にする", () => {
    render(<Field errorMessage="使えない文字があります" />);
    expect(screen.getByText("使えない文字があります")).toBeInTheDocument();
    expect(screen.queryByText("説明文")).not.toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "名前" })).toHaveAttribute("aria-invalid", "true");
  });
});
