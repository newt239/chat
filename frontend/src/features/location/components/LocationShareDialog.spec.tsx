import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { LocationShareDialog } from "./LocationShareDialog";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type Coordinates = { latitude: number; longitude: number; accuracy: number };
type Success = (position: { coords: Coordinates }) => void;
type Failure = (error: { code: number; PERMISSION_DENIED: number }) => void;

// jsdom には Geolocation API がないため、respond で成功・失敗を返す実装を navigator に生やす
const stubGeolocation = (respond: (success: Success, failure: Failure) => void) => {
  const getCurrentPosition = vi.fn(respond);
  Object.defineProperty(navigator, "geolocation", {
    configurable: true,
    value: { getCurrentPosition },
  });
  return getCurrentPosition;
};

describe("LocationShareDialog", () => {
  test("取得した現在地を確かめ、ラベルを付けて共有して閉じる", async () => {
    stubGeolocation((success) => {
      success({ coords: { accuracy: 20, latitude: 35.68, longitude: 139.76 } });
    });
    const onConfirm = vi.fn<(location: MessageLocation) => void>();
    const onClose = vi.fn<() => void>();
    render(<LocationShareDialog onConfirm={onConfirm} onClose={onClose} />);

    expect(screen.getByRole("dialog", { name: "現在地を共有" })).toBeInTheDocument();
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
    expect(onClose).toHaveBeenCalled();
  });

  test("許可されなければ理由を出し、取り直せるまで共有できない", async () => {
    const getCurrentPosition = stubGeolocation((_, failure) => {
      failure({ PERMISSION_DENIED: 1, code: 1 });
    });
    render(
      <LocationShareDialog
        onConfirm={vi.fn<(location: MessageLocation) => void>()}
        onClose={vi.fn<() => void>()}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("位置情報の利用が許可されていません");
    expect(screen.getByRole("button", { name: "共有する" })).toBeDisabled();

    await userEvent.click(screen.getByRole("button", { name: "もう一度取得" }));
    expect(getCurrentPosition).toHaveBeenCalledTimes(2);
  });
});
