import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { stubGeolocation } from "#/test/stubGeolocation";

import { LocationSharePicker } from "./LocationSharePicker";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

vi.mock("#/features/location/components/LocationMap", () => ({
  LocationMap: () => <div data-testid="map" />,
}));

describe("LocationSharePicker", () => {
  test("取得した現在地を地図で確かめ、ラベルを付けて共有する", async () => {
    stubGeolocation((success) => {
      success({ coords: { accuracy: 20, latitude: 35.68, longitude: 139.76 } });
    });
    const onConfirm = vi.fn<(location: MessageLocation) => void>();
    render(<LocationSharePicker onConfirm={onConfirm} onCancel={vi.fn<() => void>()} />);

    expect(screen.getByTestId("map")).toBeInTheDocument();
    expect(screen.getByText("35.68000, 139.76000 · 誤差 約 20 m")).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText("ラベル（任意）"), " 正面入口 ");
    await userEvent.click(screen.getByRole("button", { name: "共有する" }));

    expect(onConfirm).toHaveBeenCalledWith(
      expect.objectContaining({
        accuracyMeters: 20,
        label: "正面入口",
        latitude: 35.68,
        longitude: 139.76,
      }),
    );
  });

  test("許可されなければ理由を出し、取り直せるまで共有できない", async () => {
    const getCurrentPosition = stubGeolocation((_, failure) => {
      failure({ PERMISSION_DENIED: 1, code: 1 });
    });
    render(
      <LocationSharePicker
        onConfirm={vi.fn<(location: MessageLocation) => void>()}
        onCancel={vi.fn<() => void>()}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("位置情報の利用が許可されていません");
    expect(screen.getByRole("button", { name: "共有する" })).toBeDisabled();

    await userEvent.click(screen.getByRole("button", { name: "もう一度取得" }));
    expect(getCurrentPosition).toHaveBeenCalledTimes(2);
  });
});
