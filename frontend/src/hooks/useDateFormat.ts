import {
  formatDate,
  formatDateTime,
  formatDateWithWeekday,
  formatFullDateTime,
  formatMonthDay,
  formatRelativeTime,
  formatTime,
  formatWeekday,
} from "@chat/i18n";
import { getLocalTimeZone } from "@internationalized/date";

import { usePreferences } from "#/hooks/usePreferences";

/** 表示言語と、プロフィールのタイムゾーン（未設定なら端末のタイムゾーン）で日時を書式化する */
export const useDateFormat = () => {
  const { locale, timezone } = usePreferences();
  const timeZone = timezone || getLocalTimeZone();
  return {
    formatDate: (date: Date) => formatDate(date, locale, timeZone),
    formatDateTime: (date: Date) => formatDateTime(date, locale, timeZone),
    formatDateWithWeekday: (date: Date) => formatDateWithWeekday(date, locale, timeZone),
    formatFullDateTime: (date: Date) => formatFullDateTime(date, locale, timeZone),
    formatMonthDay: (date: Date) => formatMonthDay(date, locale, timeZone),
    formatRelativeTime: (date: Date, now: Date) => formatRelativeTime(date, now, locale),
    formatTime: (date: Date) => formatTime(date, locale, timeZone),
    formatWeekday: (date: Date) => formatWeekday(date, locale, timeZone),
    locale,
    timeZone,
  };
};
