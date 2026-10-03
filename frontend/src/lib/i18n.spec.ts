import { describe, expect, test } from "vite-plus/test";

import { i18n } from "./i18n";

describe("i18n", () => {
  test("英語は件数で単数形と複数形を切り替え、日本語は同じ文言にする", async () => {
    await i18n.changeLanguage("en");
    expect(i18n.t("channel.browse.count", { count: 1 })).toBe("1 channel");
    expect(i18n.t("channel.browse.count", { count: 2 })).toBe("2 channels");
    expect(i18n.t("thread.card.showMore", { count: 1 })).toBe("Show 1 more reply");

    await i18n.changeLanguage("ja");
    expect(i18n.t("channel.browse.count", { count: 1 })).toBe("1 件");
    expect(i18n.t("channel.browse.count", { count: 2 })).toBe("2 件");
  });
});
