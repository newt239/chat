import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChannelCategoryService } from "#/gen/chat/v1/channel_category_service_pb";
import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { NextUnreadButton } from "./NextUnreadButton";

const renderButton = (unreadCount: number) =>
  renderWithProviders(<NextUnreadButton workspaceId="ws1" />, "/app/ws1/a", (routes) => {
    routes.rpc(ChannelService.method.listChannels, () => ({
      channels: [
        create(ChannelSchema, { id: "a", isMember: true, name: "a" }),
        create(ChannelSchema, { id: "b", isMember: true, name: "b", unreadCount }),
      ],
    }));
    routes.rpc(DirectMessageService.method.listDirectMessages, () => ({ directMessages: [] }));
    routes.rpc(ChannelCategoryService.method.listChannelCategories, () => ({ categories: [] }));
  });

describe("NextUnreadButton", () => {
  test("未読のチャンネルへ移る", async () => {
    const { router } = await renderButton(2);

    await userEvent.click(await screen.findByRole("button", { name: /次の未読/ }));

    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/b");
    });
  });

  test("未読がなければ表示しない", async () => {
    await renderButton(0);

    await waitFor(() => {
      expect(screen.queryByRole("button")).toBeNull();
    });
  });
});
