import { colorModePreferences, defaultTheme, sidebarStyles } from "@chat/design-tokens";
import { defaultLocale, locales } from "@chat/i18n";
import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";
import { z } from "zod";

export const notificationLevels = ["all", "mentions", "none"] as const;

const preferencesSchema = z.object({
  locale: z.enum(locales),
  mode: z.enum(colorModePreferences),
  // 以前は端末ごとに持っていたため、端末の保存値にないことがある
  notificationLevel: z.enum(notificationLevels).default("mentions"),
  theme: z.object({
    chroma: z.number().min(0).max(0.37),
    hue: z.number().min(0).max(360),
    sidebar: z.enum(sidebarStyles),
  }),
  // IANA のタイムゾーン名。空は未設定
  timezone: z.string().default(""),
  timezoneAutoUpdate: z.boolean().default(false),
});

export type Preferences = z.infer<typeof preferencesSchema>;

export const defaultPreferences: Preferences = {
  locale: defaultLocale,
  mode: "system",
  notificationLevel: "mentions",
  theme: defaultTheme,
  timezone: "",
  timezoneAutoUpdate: false,
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
