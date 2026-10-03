import { describe, expect, test } from "vite-plus/test";

import { jumpDateSchema, jumpPresets, startOfDateKey, toDateKey } from "./dateJump";

describe("日付ジャンプ", () => {
  test("指定したタイムゾーンの日付と 0 時を相互に変換する", () => {
    const justAfterMidnight = new Date("2026-09-01T00:30:00+09:00");
    expect(toDateKey(justAfterMidnight, "Asia/Tokyo")).toBe("2026-09-01");
    expect(toDateKey(justAfterMidnight, "America/New_York")).toBe("2026-08-31");
    expect(startOfDateKey("2026-09-01", "Asia/Tokyo")).toStrictEqual(
      new Date("2026-09-01T00:00:00+09:00"),
    );
    expect(startOfDateKey("first", "Asia/Tokyo")).toStrictEqual(new Date(0));
  });

  test("今日・昨日・先週・先月の日付を求める", () => {
    expect(jumpPresets(new Date("2026-03-31T10:00:00+09:00"), "Asia/Tokyo")).toStrictEqual({
      lastMonth: "2026-02-28",
      lastWeek: "2026-03-24",
      today: "2026-03-31",
      yesterday: "2026-03-30",
    });
  });

  test("日付か first だけを受け付ける", () => {
    expect(jumpDateSchema.safeParse("2026-09-01").success).toBe(true);
    expect(jumpDateSchema.safeParse("first").success).toBe(true);
    expect(jumpDateSchema.safeParse("2026-13-01").success).toBe(false);
  });
});
