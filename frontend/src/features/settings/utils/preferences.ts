import { create } from "@bufbuild/protobuf";
import { colorModePreferences, defaultTheme, sidebarStyles } from "@chat/design-tokens";
import { resolveLocale } from "@chat/i18n";

import { ColorMode, SidebarStyle, UserPreferencesSchema } from "#/gen/chat/v1/user_pb";

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

export const preferencesToProto = ({ locale, mode, theme }: Preferences) =>
  create(UserPreferencesSchema, {
    colorMode: colorModeValues[mode],
    locale,
    theme: {
      chroma: theme.chroma,
      hue: Math.round(theme.hue) % 360,
      sidebar: sidebarStyleValues[theme.sidebar],
    },
  });

export const preferencesFromProto = ({ colorMode, locale, theme }: UserPreferences) => ({
  locale: resolveLocale(locale),
  mode: colorModePreferences.find((mode) => colorModeValues[mode] === colorMode) ?? "system",
  theme: theme
    ? {
        chroma: theme.chroma,
        hue: theme.hue,
        sidebar:
          sidebarStyles.find((style) => sidebarStyleValues[style] === theme.sidebar) ??
          defaultTheme.sidebar,
      }
    : defaultTheme,
});
