import { describe, expect, it } from "vite-plus/test";

import { createI18n, resolveLocale } from "./i18n";
import { en } from "./locales/en";
import { ja } from "./locales/ja";

const collectKeys = (messages: object, prefix = ""): string[] =>
  Object.entries(messages).flatMap(([key, value]) =>
    typeof value === "string" ? [`${prefix}${key}`] : collectKeys(value, `${prefix}${key}.`),
  );

describe("createI18n", () => {
  it("指定した言語で翻訳する", () => {
    expect(createI18n("ja", []).t("common.cancel")).toBe("キャンセル");
    expect(createI18n("en", []).t("common.cancel")).toBe("Cancel");
  });

  it("変数を埋め込める", () => {
    expect(createI18n("ja", []).t("ui.avatar.groupMembers", { count: 3 })).toBe("3 人のグループ");
  });

  it("言語を切り替えられる", async () => {
    const i18n = createI18n("ja", []);
    await i18n.changeLanguage("en");
    expect(i18n.t("common.close")).toBe("Close");
  });
});

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
