import { formatWeekday } from "@chat/i18n";

import type { HeatmapCell } from "#/gen/chat/v1/insight_service_pb";

import type { Locale } from "@chat/i18n";

// 軸の最大値を 1・2・5 の倍数に切り上げる
export const niceMax = (value: number) => {
  if (value <= 0) {
    return 1;
  }
  const power = 10 ** Math.floor(Math.log10(value));
  const normalized = value / power;
  const step = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10;
  return step * power;
};

export type Direction = "up" | "down" | "flat";

export const directionOf = (difference: number): Direction =>
  difference > 0 ? "up" : difference < 0 ? "down" : "flat";

// 前期が 0 のときは比率を出せないため null
export const percentChange = (current: number, previous: number) =>
  previous === 0 ? null : Math.round(((current - previous) / previous) * 100);

export const signed = (value: string, difference: number) => (difference > 0 ? `+${value}` : value);

// ISO の曜日（1 = 月曜日）× 24 時間の二次元配列にする
export const toHeatmapGrid = (cells: readonly HeatmapCell[]) => {
  const grid = Array.from({ length: 7 }, () => Array.from({ length: 24 }, () => 0));
  for (const cell of cells) {
    const row = grid[cell.weekday - 1];
    if (row !== undefined && cell.hour >= 0 && cell.hour < 24) {
      row[cell.hour] = cell.messageCount;
    }
  }
  return grid;
};

// 2024-01-01 は月曜日なので、ISO の曜日をそのまま日付に足せる
export const isoWeekdayLabel = (weekday: number, locale: Locale) =>
  formatWeekday(new Date(2024, 0, weekday), locale, undefined);

// サーバーが指定したタイムゾーンで区切った日付 (YYYY-MM-DD) をローカルの日付として読む
export const parseLocalDate = (value: string) => {
  const [year = 1970, month = 1, day = 1] = value.split("-").map(Number);
  return new Date(year, month - 1, day);
};
