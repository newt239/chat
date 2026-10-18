import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import {
  DirectMessageMemberSchema,
  DirectMessageSchema,
  DirectMessageService,
} from "#/gen/chat/v1/direct_message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { QuickSwitcher } from "./QuickSwitcher";

import type { ConnectRouter } from "@connectrpc/connect";

const routes = (router: ConnectRouter) => {
  router.rpc(ChannelService.method.listChannels, () => ({
    channels: [
      create(ChannelSchema, { id: "c-front", isMember: true, name: "dev/frontend" }),
      create(ChannelSchema, { id: "c-random", isMember: true, name: "random" }),
    ],
  }));
  router.rpc(DirectMessageService.method.listDirectMessages, () => ({
    directMessages: [
      create(DirectMessageSchema, {
        id: "d-bob",
        members: [create(DirectMessageMemberSchema, { displayName: "Bob", userId: "u-bob" })],
      }),
    ],
  }));
};

const renderSwitcher = () =>
  renderWithProviders(<QuickSwitcher workspaceId="ws1" onClose={() => {}} />, "/app/ws1", routes);

describe("QuickSwitcher", () => {
  test("入力でチャンネルと DM を絞り込む", async () => {
    await renderSwitcher();
    expect(await screen.findByRole("menuitem", { name: /random/ })).toBeInTheDocument();

    await userEvent.type(screen.getByRole("searchbox"), "bob");

    expect(screen.getByRole("menuitem", { name: "Bob" })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /random/ })).toBeNull();
  });

  test("選んだ会話へ移る", async () => {
    const { router } = await renderSwitcher();

    await userEvent.click(await screen.findByRole("menuitem", { name: /frontend/ }));

    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/c-front");
    });
  });

  test("入力した語でメッセージを検索できる", async () => {
    const { router } = await renderSwitcher();

    await userEvent.type(await screen.findByRole("searchbox"), "リリース");
    await userEvent.click(screen.getByRole("menuitem", { name: "「リリース」をメッセージで検索" }));

    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/search");
    });
    expect(router.state.location.search).toMatchObject({ q: "リリース" });
  });
});
