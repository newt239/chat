import { describe, expect, it } from "vite-plus/test";

import {
  formatBytes,
  formatDate,
  formatDateTime,
  formatMonthDay,
  formatNumber,
  formatRelativeTime,
  formatTime,
  formatWeekday,
} from "./format";

// ローカルタイムで組み立てるため実行環境のタイムゾーンに依存しない
const date = new Date(2026, 8, 28, 10, 16);

describe("日付・時刻のフォーマット", () => {
  it("日付を言語に合わせて表示する", () => {
    expect(formatDate(date, "ja")).toBe("2026年9月28日");
    expect(formatDate(date, "en")).toBe("Sep 28, 2026");
  });

  it("時刻を言語に合わせて表示する", () => {
    expect(formatTime(date, "ja")).toBe("10:16");
    expect(formatTime(date, "en")).toBe("10:16 AM");
  });

  it("日時を言語に合わせて表示する", () => {
    expect(formatDateTime(date, "ja")).toBe("2026年9月28日 10:16");
    expect(formatDateTime(date, "en")).toBe("Sep 28, 2026, 10:16 AM");
  });

  it("曜日を言語に合わせて表示する", () => {
    expect(formatWeekday(date, "ja")).toBe("月");
    expect(formatWeekday(date, "en")).toBe("Mon");
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
  it("月日を言語に合わせて表示する", () => {
    expect(formatMonthDay(date, "ja")).toBe("9/28");
    expect(formatMonthDay(date, "en")).toBe("9/28");
  });

  it("桁区切りを付け、小数は 1 桁までにする", () => {
    expect(formatNumber(12345.67, "ja")).toBe("12,345.7");
  });

  it("バイト数を単位付きで表示する", () => {
    expect(formatBytes(0, "ja")).toBe("0 B");
    expect(formatBytes(1536, "ja")).toBe("1.5 KB");
    expect(formatBytes(5 * 1024 ** 3, "en")).toBe("5 GB");
  });
});
