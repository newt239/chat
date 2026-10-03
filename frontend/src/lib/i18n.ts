import { defaultLocale, resolveLocale, resources } from "@chat/i18n/i18n";
import { createInstance } from "i18next";
import { initReactI18next } from "react-i18next";

import { storedPreferencesAtom } from "#/providers/store/preferences";
import { store } from "#/providers/store/store";

export const i18n = createInstance();

// 辞書は同梱しているため同期的に初期化する
void i18n.use(initReactI18next).init({
  fallbackLng: defaultLocale,
  initAsync: false,
  interpolation: { escapeValue: false },
  lng: resolveLocale(store.get(storedPreferencesAtom).locale),
  resources,
});
