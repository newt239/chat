import { create } from "@bufbuild/protobuf";
import { describe, expect, test } from "vite-plus/test";

import { MessageLocationSchema } from "#/gen/chat/v1/message_pb";
import { i18n } from "#/lib/i18n";

import { externalMapUrl, formatCoordinates } from "./location";

const location = create(MessageLocationSchema, { latitude: 35.6812, longitude: 139.7671 });

describe("externalMapUrl", () => {
  test("緯度・経度を地図アプリの検索 URL にする", () => {
    expect(externalMapUrl(location)).toBe(
      "https://www.google.com/maps/search/?api=1&query=35.6812,139.7671",
    );
  });
});

describe("formatCoordinates", () => {
  test("小数第 5 位までで表示し、誤差が分かれば添える", () => {
    expect(formatCoordinates(location, i18n.t)).toBe("35.68120, 139.76710");
    expect(formatCoordinates({ ...location, accuracyMeters: 12.4 }, i18n.t)).toBe(
      "35.68120, 139.76710 · 誤差 約 12 m",
    );
  });
});
