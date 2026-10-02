import { colorModePreferences, defaultTheme, sidebarStyles } from "@chat/design-tokens";
import { defaultLocale, locales } from "@chat/i18n";
import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";
import { z } from "zod";

export const notificationLevels = ["all", "mentions", "none"] as const;
export const channelSortOrders = ["default", "recentActivity"] as const;

const preferencesSchema = z.object({
  channelSortOrder: z.enum(channelSortOrders).default("default"),
  hideJoinMessages: z.boolean().default(false),
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
  channelSortOrder: "default",
  hideJoinMessages: false,
  locale: defaultLocale,
  mode: "system",
  notificationLevel: "mentions",
  theme: defaultTheme,
  timezone: "",
  timezoneAutoUpdate: false,
};

const preferencesStorageAtom = atomWithStorage<Preferences>(
  "preferences",
  defaultPreferences,
  undefined,
  { getOnInit: true },
);

// アカウントの設定の写し。ログイン前や GetMe の応答前も前回の設定で描画するために端末に残す。読むときは usePreferences を使う
export const storedPreferencesAtom = atom(
  (get) => preferencesSchema.safeParse(get(preferencesStorageAtom)).data ?? defaultPreferences,
  (_get, set, update: Preferences) => {
    set(preferencesStorageAtom, update);
  },
);
