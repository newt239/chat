import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { describe, expect, test } from "vite-plus/test";

import { TimelineItemSchema } from "#/gen/chat/v1/message_pb";

import { groupByDate, jumpDateSchema, jumpPresets, startOfDateKey, toDateKey } from "./dateJump";

const item = (date: Date) => create(TimelineItemSchema, { createdAt: timestampFromDate(date) });

describe("日付ジャンプ", () => {
  test("端末のタイムゾーンの日付と 0 時を相互に変換する", () => {
    expect(toDateKey(new Date(2026, 8, 1, 23, 59))).toBe("2026-09-01");
    expect(startOfDateKey("2026-09-01")).toStrictEqual(new Date(2026, 8, 1));
    expect(startOfDateKey("first")).toStrictEqual(new Date(0));
  });

  test("今日・昨日・先週・先月の日付を求める", () => {
    expect(jumpPresets(new Date(2026, 2, 31, 10))).toStrictEqual({
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

  test("古い順の項目を日付ごとにまとめる", () => {
    const groups = groupByDate([
      item(new Date(2026, 8, 1, 9)),
      item(new Date(2026, 8, 1, 18)),
      item(new Date(2026, 8, 3, 8)),
    ]);

    expect(groups.map(({ dateKey, items }) => [dateKey, items.length])).toStrictEqual([
      ["2026-09-01", 2],
      ["2026-09-03", 1],
    ]);
  });
});
