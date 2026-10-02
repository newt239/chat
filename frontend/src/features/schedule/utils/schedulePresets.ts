import { fromDate, getDayOfWeek, toCalendarDate } from "@internationalized/date";

import { atHourDaysLater } from "#/lib/zonedDate";

type SchedulePreset = "inOneHour" | "tomorrowMorning" | "nextMonday";

// 送信メニューの選択肢。日付はプロフィールのタイムゾーンで数え、月曜は今日が月曜でも翌週にする
export const schedulePresets = (now: Date, timeZone: string) => {
  const inOneHour = new Date(now.getTime() + 60 * 60 * 1000);
  inOneHour.setSeconds(0, 0);
  // en-US では日曜が 0
  const dayOfWeek = getDayOfWeek(toCalendarDate(fromDate(now, timeZone)), "en-US");
  const daysToMonday = (8 - dayOfWeek) % 7 || 7;
  const presets: { key: SchedulePreset; date: Date }[] = [
    { date: inOneHour, key: "inOneHour" },
    { date: atHourDaysLater(now, timeZone, { days: 1, hour: 9 }), key: "tomorrowMorning" },
    { date: atHourDaysLater(now, timeZone, { days: daysToMonday, hour: 9 }), key: "nextMonday" },
  ];
  return presets;
};

// 日時を指定するときの初期値
export const defaultScheduleDate = (now: Date, timeZone: string) =>
  atHourDaysLater(now, timeZone, { days: 1, hour: 9 });
