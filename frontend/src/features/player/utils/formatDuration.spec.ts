import { describe, expect, test } from "vite-plus/test";

import { formatDuration } from "./formatDuration";

describe("formatDuration", () => {
  test.each([
    [0, "0:00"],
    [5.9, "0:05"],
    [65, "1:05"],
    [3723, "1:02:03"],
    [Number.NaN, "0:00"],
    [-3, "0:00"],
  ])("%s 秒は %s", (seconds, expected) => {
    expect(formatDuration(seconds)).toBe(expected);
  });
});
