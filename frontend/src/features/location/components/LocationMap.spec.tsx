import type { ReactNode } from "react";

import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageLocationSchema } from "#/gen/chat/v1/message_pb";

import { LocationMap } from "./LocationMap";

type Point = { center: [number, number] };

// jsdom では Leaflet が描画できないため、渡された値だけを出す部品に差し替える
vi.mock("react-leaflet", () => ({
  Circle: ({ radius }: Point & { radius: number }) => (
    <div data-testid="accuracy" data-radius={radius} />
  ),
  CircleMarker: ({ center }: Point) => <div data-testid="marker" data-center={center.join(",")} />,
  MapContainer: ({ zoom, children }: { zoom: number; children: ReactNode }) => (
    <div data-testid="map" data-zoom={zoom}>
      {children}
    </div>
  ),
  TileLayer: ({ url }: { url: string }) => <div data-testid="tiles" data-url={url} />,
}));

describe("LocationMap", () => {
  test("OpenStreetMap のタイルに地点と誤差の円を描く", () => {
    render(
      <LocationMap
        location={create(MessageLocationSchema, {
          accuracyMeters: 30,
          latitude: 35.68,
          longitude: 139.76,
        })}
        className="h-40"
      />,
    );

    expect(screen.getByTestId("tiles")).toHaveAttribute(
      "data-url",
      "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
    );
    expect(screen.getByTestId("marker")).toHaveAttribute("data-center", "35.68,139.76");
    expect(screen.getByTestId("accuracy")).toHaveAttribute("data-radius", "30");
    expect(screen.getByTestId("map")).toHaveAttribute("data-zoom", "14");
  });

  test("誤差が分からなければ円を描かない", () => {
    render(
      <LocationMap
        location={create(MessageLocationSchema, { latitude: 35.68, longitude: 139.76 })}
        className="h-40"
      />,
    );

    expect(screen.queryByTestId("accuracy")).not.toBeInTheDocument();
    expect(screen.getByTestId("map")).toHaveAttribute("data-zoom", "16");
  });
});
