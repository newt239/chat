import { buildTokens, defaultTheme } from "@chat/design-tokens";
import { expect, test } from "vite-plus/test";

import { pwaColors } from "./pwaColors";

test("manifest の色が既定テーマのトークンと一致する", () => {
  const tokens = buildTokens(defaultTheme, "light");
  expect(pwaColors).toEqual({ background: tokens.bg, theme: tokens.surface });
});
