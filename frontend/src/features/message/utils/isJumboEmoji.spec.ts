import { describe, expect, test } from "vite-plus/test";

import { isJumboEmoji } from "./isJumboEmoji";

describe("isJumboEmoji", () => {
  test.each(["🎉", "👍 🙏", "👨‍👩‍👧", "👍🏽", "🇯🇵", "❤️", "🎉🎉🎉🎉🎉🎉🎉🎉"])(
    "%s は大きく表示する",
    (body) => {
      expect(isJumboEmoji(body)).toBe(true);
    },
  );

  test.each(["", "   ", "了解 👍", "123", "🎉🎉🎉🎉🎉🎉🎉🎉🎉", "**🎉**"])(
    "%s は通常の大きさ",
    (body) => {
      expect(isJumboEmoji(body)).toBe(false);
    },
  );

  test("登録済みのカスタム絵文字だけなら大きく表示する", () => {
    const names = new Map([["party", {}]]);
    expect(isJumboEmoji(":party: 🎉", names)).toBe(true);
    expect(isJumboEmoji(":unknown:", names)).toBe(false);
    expect(isJumboEmoji("ok :party:", names)).toBe(false);
  });
});
