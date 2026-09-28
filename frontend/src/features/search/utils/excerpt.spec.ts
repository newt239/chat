import { describe, expect, test } from "vite-plus/test";

import { excerpt } from "./excerpt";

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
