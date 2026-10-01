import { oklchToHex } from "./color";

export const sidebarStyles = ["tinted", "light"] as const;
export type SidebarStyle = (typeof sidebarStyles)[number];

// カスタムテーマとして保存するのはこの 3 値だけ
export type ThemeInput = {
  hue: number;
  chroma: number;
  sidebar: SidebarStyle;
};

export type ColorMode = "light" | "dark";
export const colorModePreferences = ["light", "dark", "system"] as const;
export type ColorModePreference = (typeof colorModePreferences)[number];

export const themePresets = {
  cobalt: { chroma: 0.17, hue: 262, sidebar: "tinted" },
  graphite: { chroma: 0.035, hue: 250, sidebar: "light" },
  jade: { chroma: 0.12, hue: 168, sidebar: "tinted" },
  plum: { chroma: 0.14, hue: 332, sidebar: "light" },
} as const satisfies Record<string, ThemeInput>;

export type ThemePresetName = keyof typeof themePresets;
export const themePresetNames = [
  "jade",
  "cobalt",
  "graphite",
  "plum",
] as const satisfies readonly ThemePresetName[];
export const defaultTheme: ThemeInput = themePresets.jade;

export const hueRange = { max: 360, min: 0 } as const;
export const chromaRange = { max: 0.37, min: 0 } as const;

export const findThemePreset = (theme: ThemeInput) =>
  themePresetNames.find((name) => {
    const preset = themePresets[name];
    return (
      preset.hue === theme.hue && preset.chroma === theme.chroma && preset.sidebar === theme.sidebar
    );
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
  "side",
  "side-fg",
  "side-muted",
  "side-strong",
  "side-hover",
  "side-active",
  "side-active-fg",
] as const;
export type ColorTokenName = (typeof colorTokenNames)[number];
export type ColorTokens = Record<ColorTokenName, string>;

type SidebarTokens = Pick<ColorTokens, Extract<ColorTokenName, `side${string}`>>;

const white = "#FFFFFF";
// 画像ビューアや動画の背景。写真の色を正しく見せるため表示モードによらず暗くする
const media = "#08090A";

const buildSidebarTokens = (
  { hue, chroma, sidebar }: ThemeInput,
  mode: ColorMode,
): SidebarTokens => {
  const neutral = (l: number, c = 0.009) => oklchToHex(l, c, hue);
  const accent = (l: number, k = 1) => oklchToHex(l, chroma * k, hue);
  const sideChroma = Math.min(chroma * 0.35, 0.04);
  if (sidebar === "tinted") {
    return mode === "dark"
      ? {
          side: oklchToHex(0.14, sideChroma * 0.8, hue),
          "side-active": accent(0.4, 0.55),
          "side-active-fg": neutral(0.99, 0.003),
          "side-fg": neutral(0.84, 0.012),
          "side-hover": oklchToHex(0.2, sideChroma, hue),
          "side-muted": neutral(0.64, 0.014),
          "side-strong": neutral(0.98, 0.004),
        }
      : {
          side: oklchToHex(0.25, sideChroma, hue),
          "side-active": accent(0.53),
          "side-active-fg": white,
          "side-fg": neutral(0.86, 0.014),
          "side-hover": oklchToHex(0.3, sideChroma * 1.1, hue),
          "side-muted": oklchToHex(0.7, sideChroma * 0.8, hue),
          "side-strong": white,
        };
  }
  return mode === "dark"
    ? {
        side: neutral(0.18),
        "side-active": accent(0.3, 0.38),
        "side-active-fg": accent(0.84, 0.8),
        "side-fg": neutral(0.84, 0.01),
        "side-hover": neutral(0.225),
        "side-muted": neutral(0.62, 0.01),
        "side-strong": neutral(0.97, 0.004),
      }
    : {
        side: neutral(0.955, 0.008),
        "side-active": accent(0.9, 0.35),
        "side-active-fg": accent(0.4),
        "side-fg": neutral(0.34, 0.014),
        "side-hover": neutral(0.92, 0.011),
        "side-muted": neutral(0.52, 0.014),
        "side-strong": neutral(0.18, 0.014),
      };
};

// テーマ入力から semantic token を hex で計算する。部品はこの semantic token だけを参照する
export const buildTokens = (theme: ThemeInput, mode: ColorMode): ColorTokens => {
  const { hue, chroma } = theme;
  const neutral = (l: number, c = 0.009) => oklchToHex(l, c, hue);
  const accent = (l: number, k = 1) => oklchToHex(l, chroma * k, hue);
  const pinChroma = Math.max(chroma, 0.08) * 0.35;
  const side = buildSidebarTokens(theme, mode);

  if (mode === "dark") {
    const danger = oklchToHex(0.68, 0.17, 25);
    return {
      ...side,
      accent: accent(0.72, 0.95),
      "accent-fg": accent(0.2, 0.3),
      "accent-hover": accent(0.78, 0.9),
      "accent-soft": accent(0.3, 0.38),
      "accent-text": accent(0.8, 0.85),
      badge: danger,
      "badge-fg": white,
      bg: neutral(0.165),
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
      "bookmark-bar": oklchToHex(0.74, 0.12, 165),
      "bookmark-bg": oklchToHex(0.24, 0.03, 165),
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
    ...side,
    accent: accent(0.53),
    "accent-fg": white,
    "accent-hover": accent(0.47),
    "accent-soft": accent(0.955, 0.28),
    "accent-text": accent(0.46),
    badge: danger,
    "badge-fg": white,
    bg: neutral(0.972, 0.005),
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
    "bookmark-bar": oklchToHex(0.7, 0.13, 165),
    "bookmark-bg": oklchToHex(0.972, 0.03, 165),
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
