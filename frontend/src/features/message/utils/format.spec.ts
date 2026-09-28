import { describe, expect, test } from "vite-plus/test";

import { applyFormat, detectActiveFormats } from "./format";

describe("applyFormat", () => {
  test("選択範囲を記号で囲み、カーソルを内側の末尾に置く", () => {
    expect(applyFormat("hello world", { end: 11, start: 6 }, "bold")).toEqual({
      cursor: 13,
      text: "hello **world**",
    });
  });

  test("選択がなければ記号だけを挿入する", () => {
    expect(applyFormat("", { end: 0, start: 0 }, "link")).toEqual({ cursor: 1, text: "[](url)" });
  });
});

describe("detectActiveFormats", () => {
  test("カーソルが太字の内側にあれば bold が有効", () => {
    const text = "a **bold** b";
    expect(detectActiveFormats(text, { end: 5, start: 5 }).bold).toBe(true);
    expect(detectActiveFormats(text, { end: 11, start: 11 }).bold).toBe(false);
  });

  test("行頭の記号から見出しやリストを判定する", () => {
    const text = "first\n- item";
    expect(detectActiveFormats(text, { end: text.length, start: text.length })).toMatchObject({
      heading: false,
      list: true,
    });
  });
});
