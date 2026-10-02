import { getLocalTimeZone } from "@internationalized/date";
import { describe, expect, test } from "vite-plus/test";

import { defaultScheduleDate, schedulePresets } from "./schedulePresets";

const dates = (now: Date) =>
  Object.fromEntries(schedulePresets(now, getLocalTimeZone()).map(({ key, date }) => [key, date]));

describe("schedulePresets", () => {
  test("1 時間後・明日の朝 9 時・次の月曜 9 時を返す", () => {
    // 2026-09-29 は火曜日
    const now = new Date(2026, 8, 29, 14, 25, 40);
    expect(dates(now)).toEqual({
      inOneHour: new Date(2026, 8, 29, 15, 25),
      nextMonday: new Date(2026, 9, 5, 9, 0),
      tomorrowMorning: new Date(2026, 8, 30, 9, 0),
    });
  });

  test("今日が月曜なら翌週の月曜にする", () => {
    expect(dates(new Date(2026, 8, 28, 8, 0)).nextMonday).toEqual(new Date(2026, 9, 5, 9, 0));
  });

  test("日曜なら翌日の月曜にする", () => {
    expect(dates(new Date(2026, 9, 4, 20, 0)).nextMonday).toEqual(new Date(2026, 9, 5, 9, 0));
  });
});

describe("defaultScheduleDate", () => {
  test("明日の朝 9 時を初期値にする", () => {
    expect(defaultScheduleDate(new Date(2026, 11, 31, 23, 0), getLocalTimeZone())).toEqual(
      new Date(2027, 0, 1, 9, 0),
    );
  });
});
