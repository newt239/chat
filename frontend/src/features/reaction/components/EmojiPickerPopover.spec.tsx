import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Button } from "react-aria-components";
import { describe, expect, test, vi } from "vite-plus/test";

import { CustomEmojiService } from "#/gen/chat/v1/custom_emoji_service_pb";
import { Permission, PermissionService } from "#/gen/chat/v1/permission_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { EmojiPickerPopover } from "./EmojiPickerPopover";

const setup = (myPermissions: Permission[]) =>
  renderWithProviders(
    <EmojiPickerPopover
      trigger={<Button>絵文字</Button>}
      onSelect={vi.fn<(emoji: string) => void>()}
    />,
    "/app/ws1",
    (routes) => {
      routes.rpc(PermissionService.method.getPermissions, () => ({ myPermissions }));
      routes.rpc(CustomEmojiService.method.listCustomEmojis, () => ({ emojis: [] }));
    },
  );

describe("EmojiPickerPopover", () => {
  test("登録の権限があればピッカーから絵文字を登録するダイアログを開ける", async () => {
    await setup([Permission.CREATE_CUSTOM_EMOJI]);
    await userEvent.click(screen.getByRole("button", { name: "絵文字" }));
    await userEvent.click(await screen.findByRole("button", { name: "絵文字を登録" }));

    expect(await screen.findByRole("dialog", { name: "絵文字を登録" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "名前" })).toBeInTheDocument();
  });

  test("権限がなければ登録ボタンを出さない", async () => {
    await setup([]);
    await userEvent.click(screen.getByRole("button", { name: "絵文字" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "絵文字を登録" })).not.toBeInTheDocument();
  });
});
