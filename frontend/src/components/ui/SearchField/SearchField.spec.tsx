import { useState } from "react";

import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { SearchField } from "./SearchField";

const Controlled = () => {
  const [value, setValue] = useState("");
  return (
    <>
      <SearchField label="メンバーを検索" value={value} onChange={setValue} />
      <output>{value}</output>
    </>
  );
};

describe("SearchField", () => {
  test("ラベルを名前とプレースホルダーに使い、入力を通知する", async () => {
    render(<Controlled />);
    const input = screen.getByRole("searchbox", { name: "メンバーを検索" });
    expect(input).toHaveAttribute("placeholder", "メンバーを検索");

    await userEvent.type(input, "bob");
    expect(screen.getByRole("status")).toHaveTextContent("bob");
  });

  test("Escape で入力を消す", async () => {
    render(<Controlled />);
    const input = screen.getByRole("searchbox", { name: "メンバーを検索" });
    await userEvent.type(input, "bob{Escape}");
    expect(input).toHaveValue("");
  });

  test("placeholder を渡すとラベルと別の文言を出す", () => {
    render(<SearchField label="検索" placeholder="キーワード" />);
    expect(screen.getByRole("searchbox", { name: "検索" })).toHaveAttribute(
      "placeholder",
      "キーワード",
    );
  });
});
