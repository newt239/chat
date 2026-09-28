import i18next from "i18next";

import { en } from "./locales/en";
import { ja } from "./locales/ja";

import type { ThirdPartyModule } from "i18next";

export const locales = ["ja", "en"] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = "ja";

export const resources = {
  en: { translation: en },
  ja: { translation: ja },
} as const;

export const resolveLocale = (value: string) => locales.find((l) => l === value) ?? defaultLocale;

// 辞書は同梱しているため同期的に初期化する。React なら initReactI18next を渡す
export const createI18n = (locale: Locale, plugins: ThirdPartyModule[]) => {
  const instance = i18next.createInstance();
  for (const plugin of plugins) {
    instance.use(plugin);
  }
  void instance.init({
    fallbackLng: defaultLocale,
    initAsync: false,
    interpolation: { escapeValue: false },
    lng: locale,
    resources,
  });
  return instance;
};
