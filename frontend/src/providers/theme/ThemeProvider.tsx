import { useLayoutEffect, useMemo } from "react";
import type { ReactNode } from "react";

import { useAtomValue } from "jotai";
import { MotionConfig } from "motion/react";
import { I18nProvider } from "react-aria-components";

import { i18n } from "#/lib/i18n";
import { themeVariables } from "#/lib/theme";
import { useMediaQuery } from "#/lib/useMediaQuery";
import { preferencesAtom } from "#/providers/store/preferences";

import { ColorModeContext } from "./colorMode";

type ThemeProviderProps = {
  children: ReactNode;
};

// テーマ・表示モード・言語の設定を DOM と各ライブラリに反映する
export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  const { locale, mode, theme } = useAtomValue(preferencesAtom);
  const prefersDark = useMediaQuery("(prefers-color-scheme: dark)");
  const colorMode = mode === "system" ? (prefersDark ? "dark" : "light") : mode;
  const variables = useMemo(() => themeVariables(theme, colorMode), [theme, colorMode]);

  useLayoutEffect(() => {
    const root = document.documentElement;
    for (const [name, value] of Object.entries(variables)) {
      root.style.setProperty(name, value);
    }
    root.dataset.mode = colorMode;
  }, [variables, colorMode]);

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
