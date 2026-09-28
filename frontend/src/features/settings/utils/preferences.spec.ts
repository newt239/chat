import { create } from "@bufbuild/protobuf";
import { themePresets } from "@chat/design-tokens";
import { describe, expect, test } from "vite-plus/test";

import { ColorMode, SidebarStyle, UserPreferencesSchema } from "#/gen/chat/v1/user_pb";

import { preferencesFromProto, preferencesToProto } from "./preferences";

import type { Preferences } from "#/providers/store/preferences";

describe("preferences と proto の変換", () => {
  test("往復しても値が変わらない", () => {
    const preferences: Preferences = { locale: "en", mode: "dark", theme: themePresets.plum };

    expect(preferencesFromProto(preferencesToProto(preferences))).toStrictEqual(preferences);
  });

  test("proto の列挙値に変換する", () => {
    const proto = preferencesToProto({ locale: "ja", mode: "system", theme: themePresets.jade });

    expect(proto.colorMode).toBe(ColorMode.SYSTEM);
    expect(proto.theme?.sidebar).toBe(SidebarStyle.TINTED);
  });

  test("色相は整数に丸めて 0〜359 に収める", () => {
    const proto = preferencesToProto({
      locale: "ja",
      mode: "light",
      theme: { chroma: 0.1, hue: 359.6, sidebar: "light" },
    });

    expect(proto.theme?.hue).toBe(0);
  });

  test("未設定や未知の値は既定値にする", () => {
    const preferences = preferencesFromProto(create(UserPreferencesSchema, { locale: "fr" }));

    expect(preferences).toStrictEqual({ locale: "ja", mode: "system", theme: themePresets.jade });
  });
});
