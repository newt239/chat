import { describe, expect, it } from "vite-plus/test";

import {
  formatBytes,
  formatDateTime,
  formatDateWithWeekday,
  formatFullDateTime,
  formatNumber,
  formatRelativeTime,
  formatTime,
} from "./format";

// ローカルタイムで組み立てるため実行環境のタイムゾーンに依存しない
const date = new Date(2026, 8, 28, 10, 16);
const localTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

describe("日付・時刻のフォーマット", () => {
  it("時刻を言語に合わせて表示する", () => {
    expect(formatTime(date, "ja", localTimeZone)).toBe("10:16");
    expect(formatTime(date, "en", localTimeZone)).toBe("10:16 AM");
  });

  it("指定したタイムゾーンの日時を表示する", () => {
    const utc = new Date(Date.UTC(2026, 8, 28, 1, 16));
    expect(formatTime(utc, "ja", "Asia/Tokyo")).toBe("10:16");
    expect(formatTime(utc, "en", "America/New_York")).toBe("9:16 PM");
    expect(formatDateTime(utc, "ja", "America/New_York")).toBe("2026年9月27日 21:16");
  });

  it("日時を言語に合わせて表示する", () => {
    expect(formatDateTime(date, "ja", localTimeZone)).toBe("2026年9月28日 10:16");
    expect(formatDateTime(date, "en", localTimeZone)).toBe("Sep 28, 2026, 10:16 AM");
  });

  it("曜日付きの日付を言語に合わせて表示する", () => {
    expect(formatDateWithWeekday(date, "ja", localTimeZone)).toBe("2026年9月28日(月)");
    expect(formatDateWithWeekday(date, "en", localTimeZone)).toBe("Mon, Sep 28, 2026");
  });

  it("曜日と秒を含む日時を言語に合わせて表示する", () => {
    const withSeconds = new Date(2026, 8, 28, 10, 16, 5);
    expect(formatFullDateTime(withSeconds, "ja", localTimeZone)).toBe("2026年9月28日(月) 10:16:05");
    expect(formatFullDateTime(withSeconds, "en", localTimeZone)).toBe(
      "Mon, Sep 28, 2026, 10:16:05 AM",
    );
  });
});

describe("formatRelativeTime", () => {
  const now = new Date(2026, 8, 28, 12, 0);
  const ago = (seconds: number) => new Date(now.getTime() - seconds * 1000);

  it.each([
    [10, "今", "now"],
    [3 * 60, "3 分前", "3 minutes ago"],
    [2 * 60 * 60, "2 時間前", "2 hours ago"],
    [24 * 60 * 60, "昨日", "yesterday"],
    [3 * 24 * 60 * 60, "3 日前", "3 days ago"],
    [14 * 24 * 60 * 60, "2 週間前", "2 weeks ago"],
    [400 * 24 * 60 * 60, "昨年", "last year"],
  ])("%i 秒前を表示する", (seconds, expectedJa, expectedEn) => {
    expect(formatRelativeTime(ago(seconds), now, "ja")).toBe(expectedJa);
    expect(formatRelativeTime(ago(seconds), now, "en")).toBe(expectedEn);
  });

  it("未来の日時も表示する", () => {
    expect(formatRelativeTime(ago(-5 * 60), now, "en")).toBe("in 5 minutes");
  });
});

describe("数値のフォーマット", () => {
  it("桁区切りを付け、小数は 1 桁までにする", () => {
    expect(formatNumber(12345.67, "ja")).toBe("12,345.7");
  });

  it("バイト数を単位付きで表示する", () => {
    expect(formatBytes(0, "ja")).toBe("0 B");
    expect(formatBytes(1536, "ja")).toBe("1.5 KB");
    expect(formatBytes(5 * 1024 ** 3, "en")).toBe("5 GB");
  });
});
