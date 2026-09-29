import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { stubGeolocation } from "#/test/stubGeolocation";

import { LocationShareDialog } from "./LocationShareDialog";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

vi.mock("#/features/location/components/LocationMap", () => ({
  LocationMap: () => <div data-testid="map" />,
}));

describe("LocationShareDialog", () => {
  test("共有すると位置情報を渡して閉じる", async () => {
    stubGeolocation((success) => {
      success({ coords: { accuracy: 5, latitude: 1, longitude: 2 } });
    });
    const onConfirm = vi.fn<(location: MessageLocation) => void>();
    const onOpenChange = vi.fn<(isOpen: boolean) => void>();
    render(<LocationShareDialog isOpen onOpenChange={onOpenChange} onConfirm={onConfirm} />);

    expect(screen.getByRole("dialog", { name: "現在地を共有" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "共有する" }));

    expect(onConfirm).toHaveBeenCalledWith(expect.objectContaining({ latitude: 1, longitude: 2 }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
