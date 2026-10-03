import { describe, expect, it } from "vite-plus/test";

import { oklchToHex } from "./color";

describe("oklchToHex", () => {
  it("白と黒を変換できる", () => {
    expect(oklchToHex(1, 0, 0)).toBe("#FFFFFF");
    expect(oklchToHex(0, 0, 0)).toBe("#000000");
  });

  it("sRGB の赤に近い値を返す", () => {
    expect(oklchToHex(0.628, 0.2577, 29.23)).toBe("#FF0000");
  });

  it("sRGB に収まらない彩度は下げて有効な hex を返す", () => {
    for (let hue = 0; hue < 360; hue += 15) {
      expect(oklchToHex(0.9, 0.4, hue)).toMatch(/^#[0-9A-F]{6}$/);
    }
  });
});
