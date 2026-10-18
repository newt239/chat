import { oklchToHex } from "./color";

// カスタムテーマとして保存するのはこの 2 値だけ
export type ThemeInput = {
  hue: number;
  chroma: number;
};

export type ColorMode = "light" | "dark";
export const colorModePreferences = ["system", "light", "dark"] as const;
export type ColorModePreference = (typeof colorModePreferences)[number];

export const themePresets = {
  cobalt: { chroma: 0.17, hue: 262 },
  ember: { chroma: 0.15, hue: 42 },
  graphite: { chroma: 0.035, hue: 250 },
  jade: { chroma: 0.12, hue: 168 },
  plum: { chroma: 0.14, hue: 332 },
} as const satisfies Record<string, ThemeInput>;

export const themePresetNames = [
  "ember",
  "jade",
  "cobalt",
  "graphite",
  "plum",
] as const satisfies readonly (keyof typeof themePresets)[];
export const defaultTheme: ThemeInput = themePresets.ember;

// proto の検証と同じ範囲。色相は 360 が 0 と同じ色になるため 359 までにする
export const hueRange = { max: 359, min: 0 } as const;
export const chromaRange = { max: 0.37, min: 0 } as const;

export const findThemePreset = (theme: ThemeInput) =>
  themePresetNames.find((name) => {
    const preset = themePresets[name];
    return preset.hue === theme.hue && preset.chroma === theme.chroma;
  });

export const colorTokenNames = [
  "bg",
  "surface",
  "raised",
  "sunken",
  "hover",
  "border",
  "border-strong",
  "text",
  "muted",
  "subtle",
  "accent",
  "accent-hover",
  "accent-fg",
  "accent-soft",
  "accent-text",
  "focus",
  "mention-bg",
  "mention-bar",
  "mention-chip",
  "mention-text",
  "pin-bg",
  "pin-bar",
  "bookmark-bg",
  "bookmark-bar",
  "danger",
  "danger-fg",
  "success",
  "badge",
  "badge-fg",
  "overlay",
  "media",
  "media-fg",
] as const;
type ColorTokenName = (typeof colorTokenNames)[number];
type ColorTokens = Record<ColorTokenName, string>;

const white = "#FFFFFF";
// 画像ビューアや動画の背景。写真の色を正しく見せるため表示モードによらず暗くする
const media = "#08090A";

// テーマ入力から semantic token を hex で計算する。部品はこの semantic token だけを参照する
export const buildTokens = ({ hue, chroma }: ThemeInput, mode: ColorMode): ColorTokens => {
  const neutral = (l: number, c = 0.009) => oklchToHex(l, c, hue);
  const accent = (l: number, k = 1) => oklchToHex(l, chroma * k, hue);
  const pinChroma = Math.max(chroma, 0.08) * 0.35;

  if (mode === "dark") {
    const danger = oklchToHex(0.68, 0.17, 25);
    return {
      accent: accent(0.72, 0.95),
      "accent-fg": accent(0.2, 0.3),
      "accent-hover": accent(0.78, 0.9),
      "accent-soft": accent(0.3, 0.38),
      "accent-text": accent(0.8, 0.85),
      badge: danger,
      "badge-fg": white,
      bg: neutral(0.13),
      border: neutral(0.285, 0.012),
      "border-strong": neutral(0.37, 0.012),
      danger,
      "danger-fg": white,
      focus: accent(0.72),
      hover: neutral(0.215),
      "mention-bar": oklchToHex(0.8, 0.14, 78),
      "mention-bg": oklchToHex(0.24, 0.035, 80),
      "mention-chip": oklchToHex(0.36, 0.07, 80),
      "mention-text": oklchToHex(0.92, 0.09, 85),
      muted: neutral(0.73, 0.01),
      media,
      "media-fg": white,
      overlay: "#0A0C0E99",
      "bookmark-bar": oklchToHex(0.72, 0.13, 250),
      "bookmark-bg": oklchToHex(0.25, 0.04, 250),
      "pin-bar": accent(0.72, 0.95),
      "pin-bg": oklchToHex(0.25, pinChroma, hue),
      raised: neutral(0.235),
      subtle: neutral(0.6, 0.01),
      success: oklchToHex(0.74, 0.15, 150),
      sunken: neutral(0.17),
      surface: neutral(0.195),
      text: neutral(0.93, 0.006),
    };
  }
  const danger = oklchToHex(0.56, 0.19, 25);
  return {
    accent: accent(0.53),
    "accent-fg": white,
    "accent-hover": accent(0.47),
    "accent-soft": accent(0.955, 0.28),
    "accent-text": accent(0.46),
    badge: danger,
    "badge-fg": white,
    bg: neutral(0.94, 0.012),
    border: neutral(0.915, 0.009),
    "border-strong": neutral(0.84, 0.011),
    danger,
    "danger-fg": white,
    focus: accent(0.6),
    hover: neutral(0.975, 0.005),
    "mention-bar": oklchToHex(0.78, 0.15, 75),
    "mention-bg": oklchToHex(0.975, 0.035, 88),
    "mention-chip": oklchToHex(0.92, 0.08, 85),
    "mention-text": oklchToHex(0.42, 0.1, 65),
    muted: neutral(0.47, 0.014),
    media,
    "media-fg": white,
    overlay: "#0A0C0E80",
    "bookmark-bar": oklchToHex(0.62, 0.15, 250),
    "bookmark-bg": oklchToHex(0.97, 0.025, 250),
    "pin-bar": accent(0.53),
    "pin-bg": oklchToHex(0.955, pinChroma, hue),
    raised: white,
    subtle: neutral(0.6, 0.012),
    success: oklchToHex(0.64, 0.15, 150),
    sunken: neutral(0.965, 0.006),
    surface: white,
    text: neutral(0.23, 0.014),
  };
};
