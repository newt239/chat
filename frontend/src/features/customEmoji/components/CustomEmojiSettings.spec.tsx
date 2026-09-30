import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { CustomEmojiService } from "#/gen/chat/v1/custom_emoji_service_pb";
import { Permission, PermissionService } from "#/gen/chat/v1/permission_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { CustomEmojiSettings } from "./CustomEmojiSettings";

import type { DeleteCustomEmojiRequest } from "#/gen/chat/v1/custom_emoji_service_pb";

const setup = (myPermissions: Permission[]) => {
  const deleted: DeleteCustomEmojiRequest[] = [];
  const rendered = renderWithProviders(
    <CustomEmojiSettings workspaceId="ws1" />,
    "/app/ws1",
    (routes) => {
      routes.rpc(CustomEmojiService.method.listCustomEmojis, () => ({
        emojis: [
          {
            canDelete: true,
            createdBy: { displayName: "Alice" },
            id: "e1",
            imageUrl: "a.png",
            name: "party",
          },
          {
            canDelete: false,
            createdBy: { displayName: "Bob" },
            id: "e2",
            imageUrl: "b.png",
            name: "tada",
          },
        ],
      }));
      routes.rpc(CustomEmojiService.method.deleteCustomEmoji, (req) => {
        deleted.push(req);
        return {};
      });
      routes.rpc(PermissionService.method.getPermissions, () => ({ myPermissions }));
    },
  );
  return { deleted, rendered };
};

describe("CustomEmojiSettings", () => {
  test("一覧を検索でき、削除できる絵文字だけ削除の操作を出す", async () => {
    const { deleted } = setup([Permission.CREATE_CUSTOM_EMOJI]);
    expect(await screen.findByText(":party:")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: ":party: を削除" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: ":tada: を削除" })).toBeNull();
    expect(await screen.findByRole("heading", { name: "絵文字を登録" })).toBeInTheDocument();

    await userEvent.type(screen.getByRole("searchbox", { name: "絵文字を検索" }), ":ta");
    expect(screen.queryByText(":party:")).toBeNull();
    expect(screen.getByText(":tada:")).toBeInTheDocument();

    await userEvent.clear(screen.getByRole("searchbox", { name: "絵文字を検索" }));
    await userEvent.click(screen.getByRole("button", { name: ":party: を削除" }));
    await userEvent.click(await screen.findByRole("button", { name: "削除" }));
    expect(deleted).toEqual([expect.objectContaining({ emojiId: "e1", workspaceId: "ws1" })]);
  });

  test("登録の権限がなければ登録フォームを出さない", async () => {
    setup([]);
    expect(await screen.findByText(":party:")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "絵文字を登録" })).toBeNull();
  });
});
