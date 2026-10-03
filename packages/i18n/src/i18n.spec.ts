import { describe, expect, it } from "vite-plus/test";

import { resolveLocale } from "./i18n";
import { en } from "./locales/en";
import { ja } from "./locales/ja";

const collectKeys = (messages: object, prefix = ""): string[] =>
  Object.entries(messages).flatMap(([key, value]) =>
    typeof value === "string" ? [`${prefix}${key}`] : collectKeys(value, `${prefix}${key}.`),
  );

describe("辞書", () => {
  it("日本語と英語のキーが一致する", () => {
    expect(collectKeys(en)).toStrictEqual(collectKeys(ja));
  });
});

describe("resolveLocale", () => {
  it("対応していない言語は日本語にする", () => {
    expect(resolveLocale("en")).toBe("en");
    expect(resolveLocale("fr")).toBe("ja");
  });
});
