import { typography } from "@chat/design-tokens/scale";
import { colorTokenNames, themePresets } from "@chat/design-tokens/theme";
import { describe, expect, test } from "vite-plus/test";

import globalsCss from "#/styles/globals.css?raw";

import { themeVariables } from "./theme";

describe("themeVariables", () => {
  const variables = themeVariables(themePresets.jade, "light");

  test("すべての色トークンを --c-* として出力する", () => {
    for (const name of colorTokenNames) {
      expect(variables[`--c-${name}`]).toMatch(/^#[0-9A-F]{6,8}$/);
    }
  });

  test("寸法系のトークンを px で出力する", () => {
    expect(variables["--r-lg"]).toBe("12px");
    expect(variables["--t-body-size"]).toBe("14px");
    expect(variables["--sh-lg"]).toBe("0 18px 50px -12px rgb(0 0 0 / 0.35)");
    expect(variables["--ff-sans"]).toContain('"IBM Plex Sans JP"');
  });

  // globals.css の @theme に書き漏らすと Tailwind のユーティリティが生成されないため
  test("出力するすべての変数を globals.css の @theme から参照している", () => {
    for (const name of Object.keys(variables)) {
      expect(globalsCss).toContain(`var(${name})`);
    }
    for (const name of Object.keys(typography)) {
      expect(globalsCss).toContain(`--text-${name}:`);
    }
  });
});
