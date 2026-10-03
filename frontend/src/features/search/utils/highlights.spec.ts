import { describe, expect, test } from "vite-plus/test";

import { excerpt, splitHighlights } from "./highlights";

describe("excerpt", () => {
  test("一致が先頭に近ければそのまま返す", () => {
    const ranges = [{ end: 7, start: 3 }];
    expect(excerpt("abc test", ranges, 5)).toEqual({ isTrimmed: false, ranges, text: "abc test" });
  });

  test("一致の少し前から切り出し、範囲をずらす", () => {
    expect(
      excerpt(
        "0123456789match rest",
        [
          { end: 15, start: 10 },
          { end: 20, start: 16 },
        ],
        3,
      ),
    ).toEqual({
      isTrimmed: true,
      ranges: [
        { end: 8, start: 3 },
        { end: 13, start: 9 },
      ],
      text: "789match rest",
    });
  });

  test("サロゲートペアの途中では切らない", () => {
    // 😀 は 2〜3 の 2 コード単位で、3 は下位サロゲート
    const result = excerpt("ab😀cdefgh", [{ end: 8, start: 5 }], 2);
    expect(result.text).toBe("cdefgh");
    expect(result.ranges).toEqual([{ end: 4, start: 1 }]);
  });
});

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
