// 数値はすべて px（Native では dp）。4pt グリッド
export const space = { 0: 0, 1: 4, 2: 8, 3: 12, 4: 16, 5: 20, 6: 24, 8: 32 } as const;

export const radius = { full: 999, lg: 12, md: 8, sm: 4, xl: 14 } as const;

export type TypographyToken = {
  fontSize: number;
  lineHeight: number;
  fontWeight: 400 | 500 | 600 | 700;
  mono: boolean;
};

export const typography = {
  body: { fontSize: 14, fontWeight: 400, lineHeight: 23, mono: false },
  "body-strong": { fontSize: 14, fontWeight: 700, lineHeight: 20, mono: false },
  caption: { fontSize: 11.5, fontWeight: 400, lineHeight: 16, mono: false },
  label: { fontSize: 12.5, fontWeight: 600, lineHeight: 18, mono: false },
  mono: { fontSize: 12.5, fontWeight: 400, lineHeight: 20, mono: true },
  title: { fontSize: 15, fontWeight: 700, lineHeight: 22, mono: false },
} as const satisfies Record<string, TypographyToken>;

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

// 黒の影。Web は box-shadow、Native は shadowOffset / shadowRadius / shadowOpacity に展開する
export type ShadowToken = {
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
