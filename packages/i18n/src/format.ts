import type { Locale } from "./i18n";

const toBcp47 = (locale: Locale) => (locale === "ja" ? "ja-JP" : "en-US");

// Intl のインスタンスは作るのが重いため、言語と形式の組ごとに使い回す
const memoize = <K, T>(cache: Map<K, T>, key: K, create: () => T) => {
  const hit = cache.get(key);
  if (hit !== undefined) {
    return hit;
  }
  const created = create();
  cache.set(key, created);
  return created;
};

const dateTimeFormats = new Map<string, Intl.DateTimeFormat>();
const dateTimeFormat = (locale: Locale, options: Intl.DateTimeFormatOptions) =>
  memoize(
    dateTimeFormats,
    `${locale}:${JSON.stringify(options)}`,
    () => new Intl.DateTimeFormat(toBcp47(locale), options),
  );

const relativeTimeFormats = new Map<Locale, Intl.RelativeTimeFormat>();
const numberFormats = new Map<Locale, Intl.NumberFormat>();

const dateOptions = { day: "numeric", month: "short", year: "numeric" } as const;
const timeOptions = { hour: "numeric", minute: "2-digit" } as const;

// 日付・時刻のフォーマッタは timeZone（IANA 名）の日時で表示する

// 10:16 / 10:16 AM
export const formatTime = (date: Date, locale: Locale, timeZone: string) =>
  dateTimeFormat(locale, { ...timeOptions, timeZone }).format(date);

// 2026年9月28日 10:16 / Sep 28, 2026, 10:16 AM
export const formatDateTime = (date: Date, locale: Locale, timeZone: string) =>
  dateTimeFormat(locale, { ...dateOptions, ...timeOptions, timeZone }).format(date);

// 2026年9月28日(月) 10:16:05 / Mon, Sep 28, 2026, 10:16:05 AM
export const formatFullDateTime = (date: Date, locale: Locale, timeZone: string) =>
  dateTimeFormat(locale, {
    ...dateOptions,
    ...timeOptions,
    second: "2-digit",
    timeZone,
    weekday: "short",
  }).format(date);

// 2026年9月28日(月) / Mon, Sep 28, 2026
export const formatDateWithWeekday = (date: Date, locale: Locale, timeZone: string) =>
  dateTimeFormat(locale, { ...dateOptions, timeZone, weekday: "short" }).format(date);

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
  const format = memoize(
    relativeTimeFormats,
    locale,
    () => new Intl.RelativeTimeFormat(toBcp47(locale), { numeric: "auto" }),
  );
  const diffSeconds = (date.getTime() - now.getTime()) / 1000;
  const matched = relativeUnits.find(({ seconds }) => Math.abs(diffSeconds) >= seconds);
  if (!matched) {
    return format.format(0, "second");
  }
  return format.format(Math.trunc(diffSeconds / matched.seconds), matched.unit);
};

// 1,234
export const formatNumber = (value: number, locale: Locale) =>
  memoize(
    numberFormats,
    locale,
    () => new Intl.NumberFormat(toBcp47(locale), { maximumFractionDigits: 1 }),
  ).format(value);

const byteUnits = ["B", "KB", "MB", "GB", "TB"] as const;

// 1.5 MB（1 KB = 1024 B）
export const formatBytes = (bytes: number, locale: Locale) => {
  const exponent = Math.min(
    byteUnits.length - 1,
    Math.max(0, Math.floor(Math.log(Math.max(bytes, 1)) / Math.log(1024))),
  );
  return `${formatNumber(bytes / 1024 ** exponent, locale)} ${byteUnits[exponent]}`;
};
