import { colorModePreferences, defaultTheme, sidebarStyles } from "@chat/design-tokens";
import { defaultLocale, locales } from "@chat/i18n";
import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";
import { z } from "zod";

const preferencesSchema = z.object({
  locale: z.enum(locales),
  mode: z.enum(colorModePreferences),
  theme: z.object({
    chroma: z.number().min(0).max(0.37),
    hue: z.number().min(0).max(360),
    sidebar: z.enum(sidebarStyles),
  }),
});

export type Preferences = z.infer<typeof preferencesSchema>;

export const defaultPreferences: Preferences = {
  locale: defaultLocale,
  mode: "system",
  theme: defaultTheme,
};

// ログイン前や GetMe の応答前も前回の設定で描画できるよう端末にも保持する
const preferencesStorageAtom = atomWithStorage<Preferences>(
  "preferences",
  defaultPreferences,
  undefined,
  { getOnInit: true },
);

export const preferencesAtom = atom(
  (get) => preferencesSchema.safeParse(get(preferencesStorageAtom)).data ?? defaultPreferences,
  (_get, set, update: Preferences) => {
    set(preferencesStorageAtom, update);
  },
);
