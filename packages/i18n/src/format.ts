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

// 相手の現地時刻など、端末と別のタイムゾーンの時刻
export const formatTimeInZone = (date: Date, locale: Locale, timeZone: string) =>
  new Intl.DateTimeFormat(toBcp47(locale), { ...timeOptions, timeZone }).format(date);

// 2026年9月28日 10:16 / Sep 28, 2026, 10:16 AM
export const formatDateTime = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), { ...dateOptions, ...timeOptions }).format(date);

// 2026年9月28日(月) 10:16:05 / Mon, Sep 28, 2026, 10:16:05 AM
export const formatFullDateTime = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), {
    ...dateOptions,
    ...timeOptions,
    second: "2-digit",
    weekday: "short",
  }).format(date);

// 2026年9月28日(月) / Mon, Sep 28, 2026
export const formatDateWithWeekday = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), { ...dateOptions, weekday: "short" }).format(date);

// 9/28（グラフの軸など幅の狭い場所に使う）
export const formatMonthDay = (date: Date, locale: Locale) =>
  new Intl.DateTimeFormat(toBcp47(locale), { day: "numeric", month: "numeric" }).format(date);

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

// 1,234
export const formatNumber = (value: number, locale: Locale) =>
  new Intl.NumberFormat(toBcp47(locale), { maximumFractionDigits: 1 }).format(value);

const byteUnits = ["B", "KB", "MB", "GB", "TB"] as const;

// 1.5 MB（1 KB = 1024 B）
export const formatBytes = (bytes: number, locale: Locale) => {
  const exponent = Math.min(
    byteUnits.length - 1,
    Math.max(0, Math.floor(Math.log(Math.max(bytes, 1)) / Math.log(1024))),
  );
  return `${formatNumber(bytes / 1024 ** exponent, locale)} ${byteUnits[exponent]}`;
};
