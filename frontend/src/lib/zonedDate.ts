import { Time, fromDate, toCalendarDate, toCalendarDateTime } from "@internationalized/date";

// timeZone での今日から days 日後の hour 時ちょうど
export const atHourDaysLater = (
  now: Date,
  timeZone: string,
  { days, hour }: { days: number; hour: number },
) =>
  toCalendarDateTime(toCalendarDate(fromDate(now, timeZone)).add({ days }), new Time(hour)).toDate(
    timeZone,
  );
