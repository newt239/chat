import { create } from "@bufbuild/protobuf";
import { describe, expect, test } from "vite-plus/test";

import { HeatmapCellSchema } from "#/gen/chat/v1/insight_service_pb";

import {
  directionOf,
  isoWeekdayLabel,
  niceMax,
  parseLocalDate,
  percentChange,
  signed,
  toHeatmapGrid,
} from "./chart";

describe("chart utils", () => {
  test.each([
    [0, 1],
    [7, 10],
    [12, 20],
    [34, 50],
    [51, 100],
    [100, 100],
  ])("niceMax(%i) は %i", (value, expected) => {
    expect(niceMax(value)).toBe(expected);
  });

  test("前期比を計算し、前期が 0 なら null にする", () => {
    expect(percentChange(120, 100)).toBe(20);
    expect(percentChange(80, 100)).toBe(-20);
    expect(percentChange(5, 0)).toBeNull();
    expect(directionOf(-1)).toBe("down");
    expect(directionOf(0)).toBe("flat");
    expect(signed("3", 3)).toBe("+3");
    expect(signed("-3", -3)).toBe("-3");
  });

  test("ヒートマップのセルを曜日 × 時刻に並べる", () => {
    const grid = toHeatmapGrid([
      create(HeatmapCellSchema, { hour: 9, messageCount: 4, weekday: 1 }),
      create(HeatmapCellSchema, { hour: 23, messageCount: 2, weekday: 7 }),
    ]);
    expect(grid).toHaveLength(7);
    expect(grid[0]?.[9]).toBe(4);
    expect(grid[6]?.[23]).toBe(2);
    expect(grid[3]?.[3]).toBe(0);
  });

  test("ISO の曜日を表示名にする", () => {
    expect(isoWeekdayLabel(1, "ja")).toBe("月");
    expect(isoWeekdayLabel(7, "en")).toBe("Sun");
  });

  test("日付の文字列をローカルの日付として読む", () => {
    const date = parseLocalDate("2026-09-28");
    expect([date.getFullYear(), date.getMonth(), date.getDate()]).toStrictEqual([2026, 8, 28]);
  });
});
