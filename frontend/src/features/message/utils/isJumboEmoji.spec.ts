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
});
