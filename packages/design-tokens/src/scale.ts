// 数値はすべて px
export const radius = { full: 999, lg: 12, md: 8, sm: 4, xl: 14, xs: 2 } as const;

type TypographyToken = {
  fontSize: number;
  lineHeight: number;
  fontWeight: 400 | 500 | 600 | 700;
};

export const typography = {
  body: { fontSize: 14, fontWeight: 400, lineHeight: 23 },
  "body-sm": { fontSize: 13, fontWeight: 400, lineHeight: 20 },
  "body-strong": { fontSize: 14, fontWeight: 700, lineHeight: 20 },
  caption: { fontSize: 11.5, fontWeight: 400, lineHeight: 16 },
  // 絵文字だけの投稿などで絵文字を大きく出す
  emoji: { fontSize: 32, fontWeight: 400, lineHeight: 40 },
  heading: { fontSize: 20, fontWeight: 700, lineHeight: 28 },
  label: { fontSize: 12.5, fontWeight: 600, lineHeight: 18 },
  mono: { fontSize: 12.5, fontWeight: 400, lineHeight: 20 },
  title: { fontSize: 15, fontWeight: 700, lineHeight: 22 },
} as const satisfies Record<string, TypographyToken>;

// 行の高さの倍率。文字サイズのトークンの行高より詰めたい・広げたいときに使う
export const lineHeight = { normal: 1.5, relaxed: 1.6, snug: 1.35, tight: 1.25 } as const;

// 字間（em）
export const letterSpacing = { tight: -0.01, wide: 0.06 } as const;

export const fontFamily = {
  mono: ["IBM Plex Mono", "ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
  sans: [
    "IBM Plex Sans JP",
    "Hiragino Sans",
    "Hiragino Kaku Gothic ProN",
    "Noto Sans JP",
    "system-ui",
    "sans-serif",
  ],
} as const;

// 黒の影。box-shadow に展開する
type ShadowToken = {
  offsetY: number;
  blur: number;
  spread: number;
  opacity: number;
};

export const shadow = {
  lg: { blur: 50, offsetY: 18, opacity: 0.35, spread: -12 },
  md: { blur: 14, offsetY: 4, opacity: 0.2, spread: -4 },
  sm: { blur: 2, offsetY: 1, opacity: 0.12, spread: 0 },
  xl: { blur: 80, offsetY: 30, opacity: 0.5, spread: -20 },
} as const satisfies Record<string, ShadowToken>;
