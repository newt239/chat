import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageLocationSchema } from "#/gen/chat/v1/message_pb";

import { MessageLocationCard } from "./MessageLocationCard";

vi.mock("#/features/location/components/LocationMap", () => ({
  LocationMap: () => <div data-testid="map" />,
}));

describe("MessageLocationCard", () => {
  test("ラベル・座標・誤差と、地図アプリで開くリンクを出す", () => {
    render(
      <MessageLocationCard
        location={create(MessageLocationSchema, {
          accuracyMeters: 12.4,
          label: "正面入口",
          latitude: 35.6812,
          longitude: 139.7671,
        })}
      />,
    );

    expect(screen.getByTestId("map")).toBeInTheDocument();
    expect(screen.getByText("正面入口")).toBeInTheDocument();
    expect(screen.getByText("35.68120, 139.76710 · 誤差 約 12 m")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: "地図アプリで開く" });
    expect(link).toHaveAttribute(
      "href",
      "https://www.google.com/maps/search/?api=1&query=35.6812,139.7671",
    );
    expect(link).toHaveAttribute("target", "_blank");
  });

  test("ラベルがなければ「位置情報」と表示する", () => {
    render(
      <MessageLocationCard
        location={create(MessageLocationSchema, { latitude: 1, longitude: 2 })}
      />,
    );

    expect(screen.getByText("位置情報")).toBeInTheDocument();
  });
});
