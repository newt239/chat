import { create } from "@bufbuild/protobuf";
import { colorModePreferences, defaultTheme, sidebarStyles } from "@chat/design-tokens";
import { resolveLocale } from "@chat/i18n";

import {
  ColorMode,
  NotificationLevel,
  SidebarStyle,
  UserPreferencesSchema,
} from "#/gen/chat/v1/user_pb";
import { notificationLevels } from "#/providers/store/preferences";

import type { UserPreferences } from "#/gen/chat/v1/user_pb";
import type { Preferences } from "#/providers/store/preferences";

import type { ColorModePreference, SidebarStyle as SidebarStyleName } from "@chat/design-tokens";

const sidebarStyleValues: Record<SidebarStyleName, SidebarStyle> = {
  light: SidebarStyle.LIGHT,
  tinted: SidebarStyle.TINTED,
};

const colorModeValues: Record<ColorModePreference, ColorMode> = {
  dark: ColorMode.DARK,
  light: ColorMode.LIGHT,
  system: ColorMode.SYSTEM,
};

const notificationLevelValues: Record<Preferences["notificationLevel"], NotificationLevel> = {
  all: NotificationLevel.ALL,
  mentions: NotificationLevel.MENTIONS,
  none: NotificationLevel.NONE,
};

export const preferencesToProto = ({
  locale,
  mode,
  notificationLevel,
  theme,
  timezone,
  timezoneAutoUpdate,
}: Preferences) =>
  create(UserPreferencesSchema, {
    colorMode: colorModeValues[mode],
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

export const preferencesFromProto = ({
  colorMode,
  locale,
  notificationLevel,
  theme,
  timezone,
  timezoneAutoUpdate,
}: UserPreferences) => ({
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
