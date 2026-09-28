import { describe, expect, test } from "vite-plus/test";

import { waveformBars } from "./waveformBars";

describe("waveformBars", () => {
  test("同じ文字列からは同じ波形を作り、高さは 0.25〜1 に収まる", () => {
    const bars = waveformBars("attachment-1");
    expect(bars).toEqual(waveformBars("attachment-1"));
    expect(bars).not.toEqual(waveformBars("attachment-2"));
    expect(bars.every((height) => height >= 0.25 && height <= 1)).toBe(true);
  });
});
