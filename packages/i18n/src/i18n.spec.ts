import { describe, expect, it } from "vite-plus/test";

import { resolveLocale } from "./i18n";

describe("resolveLocale", () => {
  it("対応していない言語は日本語にする", () => {
    expect(resolveLocale("en")).toBe("en");
    expect(resolveLocale("fr")).toBe("ja");
  });
});
