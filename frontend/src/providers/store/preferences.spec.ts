import { createStore } from "jotai";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

// atomWithStorage は atom の生成時に端末の値を読むため、保存してからモジュールを読み込み直す
const loadPreferences = async (saved: object) => {
  localStorage.setItem("preferences", JSON.stringify(saved));
  vi.resetModules();
  const { defaultPreferences, preferencesAtom } = await import("./preferences");
  return { defaultPreferences, preferences: createStore().get(preferencesAtom) };
};

describe("preferencesAtom", () => {
  afterEach(() => {
    localStorage.clear();
  });

  test("端末に保存した設定を読み込む", async () => {
    const saved = {
      locale: "en",
      mode: "dark",
      theme: { chroma: 0.2, hue: 100, sidebar: "light" },
    };

    const { preferences } = await loadPreferences(saved);

    expect(preferences).toStrictEqual(saved);
  });

  test("壊れた値は既定値に置き換える", async () => {
    const { defaultPreferences, preferences } = await loadPreferences({
      locale: "fr",
      mode: "dark",
    });

    expect(preferences).toStrictEqual(defaultPreferences);
  });
});
