import { describe, expect, test } from "vite-plus/test";

import { atHourDaysLater } from "./zonedDate";

describe("atHourDaysLater", () => {
  test("指定したタイムゾーンの日付で数える", () => {
    // 東京では 10/2 の 0:30、ニューヨークではまだ 10/1
    const now = new Date("2026-10-01T15:30:00Z");
    expect(atHourDaysLater(now, "Asia/Tokyo", { days: 1, hour: 9 })).toEqual(
      new Date("2026-10-03T09:00:00+09:00"),
    );
    expect(atHourDaysLater(now, "America/New_York", { days: 1, hour: 9 })).toEqual(
      new Date("2026-10-02T09:00:00-04:00"),
    );
  });
});
