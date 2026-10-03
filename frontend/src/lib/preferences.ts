import { create } from "@bufbuild/protobuf";
import { colorModePreferences, defaultTheme, sidebarStyles } from "@chat/design-tokens/theme";
import { resolveLocale } from "@chat/i18n/i18n";

import {
  ChannelSortOrder,
  ColorMode,
  NotificationLevel,
  SidebarStyle,
  UserPreferencesSchema,
} from "#/gen/chat/v1/user_pb";
import { channelSortOrders, notificationLevels } from "#/providers/store/preferences";

import type { UserPreferences } from "#/gen/chat/v1/user_pb";
import type { Preferences } from "#/providers/store/preferences";

import type {
  ColorModePreference,
  SidebarStyle as SidebarStyleName,
} from "@chat/design-tokens/theme";

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

const channelSortOrderValues: Record<Preferences["channelSortOrder"], ChannelSortOrder> = {
  default: ChannelSortOrder.DEFAULT,
  recentActivity: ChannelSortOrder.RECENT_ACTIVITY,
};

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
