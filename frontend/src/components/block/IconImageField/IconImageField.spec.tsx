import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { ImagePurpose } from "#/gen/chat/v1/image_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { IconImageField } from "./IconImageField";

const setup = (value: string) => {
  const onChange = vi.fn<(url: string) => void>();
  return renderWithProviders(
    <IconImageField
      label="アイコン"
      name="Alice"
      value={value}
      onChange={onChange}
      purpose={ImagePurpose.AVATAR}
      workspaceId={null}
    />,
    "/app/ws1",
    () => {},
  ).then(() => onChange);
};

describe("IconImageField", () => {
  beforeEach(() => {
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: () => "blob:icon" });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: () => {} });
  });

  test("設定済みならリセットで空にできる", async () => {
    const onChange = await setup("https://example.com/a.png");
    expect(screen.getByRole("group", { name: "アイコン" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "リセット" }));
    expect(onChange).toHaveBeenCalledWith("");
  });

  test("未設定ならリセットを出さず、画像を選ぶと切り抜きのダイアログを開く", async () => {
    await setup("");
    expect(screen.queryByRole("button", { name: "リセット" })).not.toBeInTheDocument();

    const input = document.querySelector<HTMLInputElement>('input[type="file"]');
    if (input === null) {
      throw new Error("file input not found");
    }
    await userEvent.upload(input, new File(["x"], "a.png", { type: "image/png" }));
    expect(await screen.findByRole("dialog", { name: "画像を切り抜く" })).toBeInTheDocument();
    expect(screen.getByRole("slider", { name: "拡大" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "画像を切り抜く" })).not.toBeInTheDocument();
    });
  });
});
