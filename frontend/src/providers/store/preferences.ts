import { create } from "@bufbuild/protobuf";
import { colorModePreferences, defaultTheme, sidebarStyles } from "@chat/design-tokens/theme";
import { resolveLocale } from "@chat/i18n/i18n";
import { atomWithStorage } from "jotai/utils";

import {
  ChannelSortOrder,
  ColorMode,
  NotificationLevel,
  SidebarStyle,
  UserPreferencesSchema,
} from "#/gen/chat/v1/user_pb";

import type { UserPreferences } from "#/gen/chat/v1/user_pb";

import type {
  ColorModePreference,
  SidebarStyle as SidebarStyleName,
} from "@chat/design-tokens/theme";

export const notificationLevels = ["all", "mentions", "none"] as const;
export const channelSortOrders = ["default", "recentActivity"] as const;

const sidebarStyleValues: Record<SidebarStyleName, SidebarStyle> = {
  light: SidebarStyle.LIGHT,
  tinted: SidebarStyle.TINTED,
};

const colorModeValues: Record<ColorModePreference, ColorMode> = {
  dark: ColorMode.DARK,
  light: ColorMode.LIGHT,
  system: ColorMode.SYSTEM,
};

const notificationLevelValues: Record<(typeof notificationLevels)[number], NotificationLevel> = {
  all: NotificationLevel.ALL,
  mentions: NotificationLevel.MENTIONS,
  none: NotificationLevel.NONE,
};

const channelSortOrderValues: Record<(typeof channelSortOrders)[number], ChannelSortOrder> = {
  default: ChannelSortOrder.DEFAULT,
  recentActivity: ChannelSortOrder.RECENT_ACTIVITY,
};

/** 未設定や未知の値は既定値にする */
export const preferencesFromProto = ({
  channelSortOrder,
  colorMode,
  locale,
  notificationLevel,
  hideJoinMessages,
  theme,
  timezone,
  timezoneAutoUpdate,
}: UserPreferences) => ({
  channelSortOrder:
    channelSortOrders.find((order) => channelSortOrderValues[order] === channelSortOrder) ??
    "default",
  hideJoinMessages,
  locale: resolveLocale(locale),
  mode: colorModePreferences.find((mode) => colorModeValues[mode] === colorMode) ?? "system",
  notificationLevel:
    notificationLevels.find((level) => notificationLevelValues[level] === notificationLevel) ??
    "mentions",
  theme: theme
    ? {
        chroma: theme.chroma,
        hue: theme.hue,
        sidebar:
          sidebarStyles.find((style) => sidebarStyleValues[style] === theme.sidebar) ??
          defaultTheme.sidebar,
      }
    : defaultTheme,
  timezone,
  timezoneAutoUpdate,
});

export type Preferences = ReturnType<typeof preferencesFromProto>;

export const preferencesToProto = ({
  channelSortOrder,
  locale,
  mode,
  notificationLevel,
  hideJoinMessages,
  theme,
  timezone,
  timezoneAutoUpdate,
}: Preferences) =>
  create(UserPreferencesSchema, {
    channelSortOrder: channelSortOrderValues[channelSortOrder],
    colorMode: colorModeValues[mode],
    hideJoinMessages,
    locale,
    notificationLevel: notificationLevelValues[notificationLevel],
    theme: {
      chroma: theme.chroma,
      hue: Math.round(theme.hue) % 360,
      sidebar: sidebarStyleValues[theme.sidebar],
    },
    timezone,
    timezoneAutoUpdate,
  });

// アカウントの設定の写し。ログイン前や GetMe の応答前も前回の設定で描画するために端末に残す。読むときは usePreferences を使う
export const storedPreferencesAtom = atomWithStorage<UserPreferences>(
  "preferences",
  create(UserPreferencesSchema),
  undefined,
  { getOnInit: true },
);
