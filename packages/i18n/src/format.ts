import type { Locale } from "./i18n";

const toBcp47 = (locale: Locale) => (locale === "ja" ? "ja-JP" : "en-US");

const dateOptions = { day: "numeric", month: "short", year: "numeric" } as const;
const timeOptions = { hour: "numeric", minute: "2-digit" } as const;

// 2026年9月28日 / Sep 28, 2026
export const formatDate = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), dateOptions).format(date);

// 10:16 / 10:16 AM
export const formatTime = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), timeOptions).format(date);

// 2026年9月28日 10:16 / Sep 28, 2026, 10:16 AM
export const formatDateTime = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), { ...dateOptions, ...timeOptions }).format(date);

// 月 / Mon
export const formatWeekday = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), { weekday: "short" }).format(date);

const relativeUnits = [
  { seconds: 60 * 60 * 24 * 365, unit: "year" },
  { seconds: 60 * 60 * 24 * 30, unit: "month" },
  { seconds: 60 * 60 * 24 * 7, unit: "week" },
  { seconds: 60 * 60 * 24, unit: "day" },
  { seconds: 60 * 60, unit: "hour" },
  { seconds: 60, unit: "minute" },
] as const satisfies readonly { seconds: number; unit: Intl.RelativeTimeFormatUnit }[];

// 3 分前 / 3 minutes ago。1 分未満は「今」
export const formatRelativeTime = (date: Date, now: Date, locale: Locale) => {
  const format = new Intl.RelativeTimeFormat(toBcp47(locale), { numeric: "auto" });
  const diffSeconds = (date.getTime() - now.getTime()) / 1000;
  const matched = relativeUnits.find(({ seconds }) => Math.abs(diffSeconds) >= seconds);
  if (!matched) {
    return format.format(0, "second");
  }
  return format.format(Math.trunc(diffSeconds / matched.seconds), matched.unit);
};
