import { create } from "@bufbuild/protobuf";
import { themePresets } from "@chat/design-tokens/theme";
import { describe, expect, test } from "vite-plus/test";

import {
  ColorMode,
  NotificationLevel,
  SidebarStyle,
  UserPreferencesSchema,
} from "#/gen/chat/v1/user_pb";

import { preferencesFromProto, preferencesToProto } from "./preferences";

import type { Preferences } from "./preferences";

const defaults = preferencesFromProto(create(UserPreferencesSchema));

describe("preferences と proto の変換", () => {
  test("往復しても値が変わらない", () => {
    const preferences: Preferences = {
      channelSortOrder: "recentActivity",
      hideJoinMessages: true,
      locale: "en",
      mode: "dark",
      notificationLevel: "all",
      theme: themePresets.plum,
      timezone: "Asia/Tokyo",
      timezoneAutoUpdate: true,
    };

    expect(preferencesFromProto(preferencesToProto(preferences))).toStrictEqual(preferences);
  });

  test("proto の列挙値に変換する", () => {
    const proto = preferencesToProto({ ...defaults, notificationLevel: "none" });

    expect(proto.colorMode).toBe(ColorMode.SYSTEM);
    expect(proto.notificationLevel).toBe(NotificationLevel.NONE);
    expect(proto.theme?.sidebar).toBe(SidebarStyle.TINTED);
  });

  test("色相は整数に丸めて 0〜359 に収める", () => {
    const proto = preferencesToProto({
      ...defaults,
      theme: { chroma: 0.1, hue: 359.6, sidebar: "light" },
    });

    expect(proto.theme?.hue).toBe(0);
  });

  test("未設定や未知の値は既定値にする", () => {
    const preferences = preferencesFromProto(create(UserPreferencesSchema, { locale: "fr" }));

    expect(preferences).toStrictEqual({
      channelSortOrder: "default",
      hideJoinMessages: false,
      locale: "ja",
      mode: "system",
      notificationLevel: "mentions",
      theme: themePresets.jade,
      timezone: "",
      timezoneAutoUpdate: false,
    });
  });
});
