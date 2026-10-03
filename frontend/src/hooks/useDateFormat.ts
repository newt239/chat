import {
  formatDateTime,
  formatDateWithWeekday,
  formatFullDateTime,
  formatTime,
} from "@chat/i18n/format";
import { getLocalTimeZone } from "@internationalized/date";

import { usePreferences } from "#/hooks/usePreferences";

/** 表示言語と、プロフィールのタイムゾーン（未設定なら端末のタイムゾーン）で日時を書式化する */
export const useDateFormat = () => {
  const { locale, timezone } = usePreferences();
  const timeZone = timezone || getLocalTimeZone();
  return {
    formatDateTime: (date: Date) => formatDateTime(date, locale, timeZone),
    formatDateWithWeekday: (date: Date) => formatDateWithWeekday(date, locale, timeZone),
    formatFullDateTime: (date: Date) => formatFullDateTime(date, locale, timeZone),
    formatTime: (date: Date) => formatTime(date, locale, timeZone),
    locale,
    timeZone,
  };
};
