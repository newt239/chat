import { describe, expect, test } from "vite-plus/test";

import { splitHighlights } from "./splitHighlights";

describe("splitHighlights", () => {
  test("一致範囲とそれ以外に分ける", () => {
    expect(splitHighlights("hello world", [{ end: 5, start: 0 }])).toEqual([
      { isMatch: true, start: 0, text: "hello" },
      { isMatch: false, start: 5, text: " world" },
    ]);
  });

  test("複数の範囲と末尾の一致を扱う", () => {
    expect(
      splitHighlights("a bc d", [
        { end: 1, start: 0 },
        { end: 6, start: 5 },
      ]),
    ).toEqual([
      { isMatch: true, start: 0, text: "a" },
      { isMatch: false, start: 1, text: " bc " },
      { isMatch: true, start: 5, text: "d" },
    ]);
  });

  test("範囲がなければ全体を 1 つにし、はみ出した範囲は切り詰める", () => {
    expect(splitHighlights("abc", [])).toEqual([{ isMatch: false, start: 0, text: "abc" }]);
    expect(splitHighlights("abc", [{ end: 10, start: 2 }])).toEqual([
      { isMatch: false, start: 0, text: "ab" },
      { isMatch: true, start: 2, text: "c" },
    ]);
  });
});
