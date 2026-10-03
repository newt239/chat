import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChannelCategoryService } from "#/gen/chat/v1/channel_category_service_pb";
import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { useNextUnreadChannel } from "./useNextUnreadChannel";

const NextButton = () => {
  const { goNext, nextId } = useNextUnreadChannel("ws1");
  return (
    <button type="button" onClick={goNext}>
      {nextId ?? "なし"}
    </button>
  );
};

describe("useNextUnreadChannel", () => {
  test("サイドバーで次にある未読のチャンネルへ移る", async () => {
    const { router } = await renderWithProviders(<NextButton />, "/app/ws1/a", (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [
          create(ChannelSchema, { id: "a", isMember: true, name: "a" }),
          create(ChannelSchema, { id: "b", isMember: true, name: "b" }),
          create(ChannelSchema, { id: "c", isMember: true, name: "c", unreadCount: 1 }),
        ],
      }));
      routes.rpc(DirectMessageService.method.listDirectMessages, () => ({ directMessages: [] }));
      routes.rpc(ChannelCategoryService.method.listChannelCategories, () => ({ categories: [] }));
    });

    await userEvent.click(await screen.findByRole("button", { name: "c" }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/c");
    });
  });
});
