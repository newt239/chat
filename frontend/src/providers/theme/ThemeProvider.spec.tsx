import { create } from "@bufbuild/protobuf";
import { buildTokens, themePresets } from "@chat/design-tokens/theme";
import { render, screen } from "@testing-library/react";
import { createStore, Provider } from "jotai";
import { useTranslation } from "react-i18next";
import { afterEach, describe, expect, test } from "vite-plus/test";

import { UserPreferencesSchema } from "#/gen/chat/v1/user_pb";
import {
  preferencesFromProto,
  preferencesToProto,
  storedPreferencesAtom,
} from "#/providers/store/preferences";
import { QueryWrapper } from "#/test/QueryWrapper";

import { useColorMode } from "./colorMode";
import { ThemeProvider } from "./ThemeProvider";

import type { Preferences } from "#/providers/store/preferences";

const Probe = () => {
  const { t } = useTranslation();
  return (
    <p>
      {useColorMode()} {t("common.cancel")}
    </p>
  );
};

const defaults = preferencesFromProto(create(UserPreferencesSchema));

const renderWithPreferences = (preferences: Partial<Preferences>) => {
  const store = createStore();
  store.set(storedPreferencesAtom, preferencesToProto({ ...defaults, ...preferences }));
  render(
    <Provider store={store}>
      <ThemeProvider>
        <Probe />
      </ThemeProvider>
    </Provider>,
    { wrapper: QueryWrapper },
  );
};

const root = document.documentElement;

describe("ThemeProvider", () => {
  afterEach(() => {
    localStorage.clear();
  });

  test("テーマのトークンを CSS 変数として書き込む", () => {
    renderWithPreferences({ mode: "light", theme: themePresets.cobalt });

    const tokens = buildTokens(themePresets.cobalt, "light");
    expect(root.style.getPropertyValue("--c-accent")).toBe(tokens.accent);
    expect(root.style.getPropertyValue("--c-side")).toBe(tokens.side);
    expect(root.style.getPropertyValue("--r-md")).toBe("8px");
    expect(root.dataset.mode).toBe("light");
  });

  test("ダークモードを反映し、子にも解決後のモードを渡す", () => {
    const themeColor = document.createElement("meta");
    themeColor.name = "theme-color";
    document.head.append(themeColor);
    renderWithPreferences({ mode: "dark" });

    const { surface } = buildTokens(themePresets.jade, "dark");
    expect(root.dataset.mode).toBe("dark");
    expect(root.style.getPropertyValue("--c-surface")).toBe(surface);
    expect(themeColor.content).toBe(surface);
    expect(screen.getByText(/dark/)).toBeInTheDocument();
  });

  test("system はOSの設定に従う（テストではライト）", () => {
    renderWithPreferences({ mode: "system" });

    expect(root.dataset.mode).toBe("light");
  });

  test("言語を切り替える", () => {
    renderWithPreferences({ locale: "en", mode: "light" });

    expect(root.lang).toBe("en");
    expect(screen.getByText(/Cancel/)).toBeInTheDocument();
  });
});
