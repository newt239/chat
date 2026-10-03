import { describe, expect, it } from "vite-plus/test";

import {
  buildTokens,
  colorTokenNames,
  findThemePreset,
  themePresetNames,
  themePresets,
} from "./theme";

import type { ColorMode, ThemeInput } from "./theme";

// WCAG のコントラスト比
const relativeLuminance = (hex: string) => {
  const [r = 0, g = 0, b = 0] = [1, 3, 5].map((i) => {
    const c = Number.parseInt(hex.slice(i, i + 2), 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};

const contrastRatio = (a: string, b: string) => {
  const [high, low] = [relativeLuminance(a), relativeLuminance(b)].toSorted((x, y) => y - x);
  return ((high ?? 0) + 0.05) / ((low ?? 0) + 0.05);
};

const modes: ColorMode[] = ["light", "dark"];
const presets = themePresetNames.map((name) => [name, themePresets[name]] as const);

describe("buildTokens", () => {
  it.each(presets)("%s は全トークンを hex で返す", (_name, theme) => {
    for (const mode of modes) {
      const tokens = buildTokens(theme, mode);
      expect(Object.keys(tokens).toSorted()).toStrictEqual([...colorTokenNames].toSorted());
      for (const value of Object.values(tokens)) {
        expect(value).toMatch(/^#[0-9A-F]{6}([0-9A-F]{2})?$/);
      }
    }
  });

  it("プリセットの色が意図せず変わらない", () => {
    expect(buildTokens(themePresets.jade, "light").accent).toBe("#0E7F60");
  });

  it.each(presets)("%s は本文と補足の文字が読めるコントラストを保つ", (_name, theme) => {
    for (const mode of modes) {
      const t = buildTokens(theme, mode);
      expect(contrastRatio(t.text, t.surface)).toBeGreaterThanOrEqual(7);
      expect(contrastRatio(t.muted, t.surface)).toBeGreaterThanOrEqual(4.5);
      expect(contrastRatio(t["accent-text"], t.surface)).toBeGreaterThanOrEqual(4.5);
      expect(contrastRatio(t["side-fg"], t.side)).toBeGreaterThanOrEqual(4.5);
      expect(contrastRatio(t["side-active-fg"], t["side-active"])).toBeGreaterThanOrEqual(4.5);
      expect(contrastRatio(t["accent-fg"], t.accent)).toBeGreaterThanOrEqual(4.5);
    }
  });

  it("任意の色相でも本文のコントラストを保つ", () => {
    for (let hue = 0; hue < 360; hue += 30) {
      for (const sidebar of ["tinted", "light"] as const) {
        const theme: ThemeInput = { chroma: 0.2, hue, sidebar };
        for (const mode of modes) {
          const t = buildTokens(theme, mode);
          expect(contrastRatio(t.text, t.surface)).toBeGreaterThanOrEqual(7);
          expect(contrastRatio(t["side-fg"], t.side)).toBeGreaterThanOrEqual(4.5);
        }
      }
    }
  });
});

describe("findThemePreset", () => {
  it("プリセットと一致すれば名前を返す", () => {
    expect(findThemePreset({ chroma: 0.17, hue: 262, sidebar: "tinted" })).toBe("cobalt");
  });

  it("一致しなければ undefined を返す", () => {
    expect(findThemePreset({ chroma: 0.17, hue: 100, sidebar: "tinted" })).toBeUndefined();
  });
});
