import { describe, expect, test } from "vite-plus/test";

import { toLocalDateTime } from "./localDateTime";

describe("toLocalDateTime", () => {
  test("ブラウザのタイムゾーンの日時を分まで 0 埋めで返し、Date で読み戻せる", () => {
    const date = new Date(2026, 0, 2, 3, 4, 59);
    expect(toLocalDateTime(date)).toBe("2026-01-02T03:04");
    expect(new Date(toLocalDateTime(date)).getTime()).toBe(new Date(2026, 0, 2, 3, 4).getTime());
  });
});
