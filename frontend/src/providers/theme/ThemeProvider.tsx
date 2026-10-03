import { useEffect, useLayoutEffect } from "react";
import type { ReactNode } from "react";

import { useSetAtom } from "jotai";
import { MotionConfig } from "motion/react";
import { I18nProvider } from "react-aria-components";

import { useMe } from "#/hooks/useMe";
import { useMediaQuery } from "#/hooks/useMediaQuery";
import { usePreferences } from "#/hooks/usePreferences";
import { i18n } from "#/lib/i18n";
import { preferencesFromProto } from "#/lib/preferences";
import { storedPreferencesAtom } from "#/providers/store/preferences";

import { ColorModeContext } from "./colorMode";
import { themeVariables } from "./theme";

type ThemeProviderProps = {
  children: ReactNode;
};

// テーマ・表示モード・言語の設定を DOM と各ライブラリに反映する
export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  const { locale, mode, theme } = usePreferences();
  const accountPreferences = useMe().data?.preferences;
  const setStoredPreferences = useSetAtom(storedPreferencesAtom);
  const prefersDark = useMediaQuery("(prefers-color-scheme: dark)");
  const colorMode = mode === "system" ? (prefersDark ? "dark" : "light") : mode;
  const variables = themeVariables(theme, colorMode);

  useLayoutEffect(() => {
    const root = document.documentElement;
    for (const [name, value] of Object.entries(variables)) {
      root.style.setProperty(name, value);
    }
    root.dataset.mode = colorMode;
    // ホーム画面から開いたときのステータスバーを見出しの色に合わせる
    document
      .querySelector('meta[name="theme-color"]')
      ?.setAttribute("content", variables["--c-surface"] ?? "");
  }, [variables, colorMode]);

  // 次に開いたときもログイン前からこの設定で描画できるよう、端末に写しを残す
  useEffect(() => {
    if (accountPreferences) {
      setStoredPreferences(preferencesFromProto(accountPreferences));
    }
  }, [accountPreferences, setStoredPreferences]);

  useLayoutEffect(() => {
    document.documentElement.lang = locale;
    void i18n.changeLanguage(locale);
  }, [locale]);

  return (
    <ColorModeContext value={colorMode}>
      <I18nProvider locale={locale}>
        <MotionConfig reducedMotion="user">{children}</MotionConfig>
      </I18nProvider>
    </ColorModeContext>
  );
};
