import { en } from "./locales/en";
import { ja } from "./locales/ja";

export const locales = ["ja", "en"] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = "ja";

export const resources = {
  en: { translation: en },
  ja: { translation: ja },
} as const;

export const resolveLocale = (value: string) => locales.find((l) => l === value) ?? defaultLocale;
