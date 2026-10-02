import { fontFamily, radius, shadow, typography } from "@chat/design-tokens/scale";
import { buildTokens } from "@chat/design-tokens/theme";

import type { ColorMode, ThemeInput } from "@chat/design-tokens/theme";

const px = (value: number) => `${value}px`;
const fontStack = (families: readonly string[]) =>
  families.map((f) => (f.includes(" ") ? `"${f}"` : f)).join(", ");

// テーマに依存しない寸法系のトークン。globals.css の @theme がこれらの変数を参照する
const staticEntries = [
  ...Object.entries(radius).map(([name, value]) => [`--r-${name}`, px(value)] as const),
  ...Object.entries(shadow).map(
    ([name, s]) =>
      [
        `--sh-${name}`,
        `0 ${px(s.offsetY)} ${px(s.blur)} ${px(s.spread)} rgb(0 0 0 / ${s.opacity})`,
      ] as const,
  ),
  ...Object.entries(typography).flatMap(([name, t]) => [
    [`--t-${name}-size`, px(t.fontSize)] as const,
    [`--t-${name}-line`, px(t.lineHeight)] as const,
    [`--t-${name}-weight`, String(t.fontWeight)] as const,
  ]),
  ["--ff-sans", fontStack(fontFamily.sans)] as const,
  ["--ff-mono", fontStack(fontFamily.mono)] as const,
];

export const themeVariables = (theme: ThemeInput, mode: ColorMode) => ({
  ...Object.fromEntries(staticEntries),
  ...Object.fromEntries(
    Object.entries(buildTokens(theme, mode)).map(([name, value]) => [`--c-${name}`, value]),
  ),
});
