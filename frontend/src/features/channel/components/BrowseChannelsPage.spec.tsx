import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";
import {
  BrowsableChannelSchema,
  ChannelSchema,
  ChannelService,
} from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { BrowseChannelsPage } from "./BrowseChannelsPage";

import type { JoinChannelRequest } from "#/gen/chat/v1/channel_member_service_pb";

const channels = [
  create(BrowsableChannelSchema, {
    channel: create(ChannelSchema, { description: "デザインの相談", id: "d", name: "design" }),
    memberCount: 3,
  }),
  create(BrowsableChannelSchema, {
    channel: create(ChannelSchema, { id: "g", isMember: true, name: "general" }),
    memberCount: 12,
  }),
];

describe("BrowseChannelsPage", () => {
  test("未参加のチャンネルに参加でき、名前や説明で絞り込める", async () => {
    const joined = vi.fn<(req: JoinChannelRequest) => void>();
    await renderWithProviders(<BrowseChannelsPage />, "/app/ws1/browse-channels", (routes) => {
      routes.rpc(ChannelService.method.listBrowsableChannels, () => ({ channels }));
      routes.rpc(ChannelMemberService.method.joinChannel, (req) => {
        joined(req);
        return {};
      });
    });

    expect(await screen.findByRole("link", { name: /design/ })).toBeInTheDocument();
    expect(screen.getByText("参加中")).toBeInTheDocument();
    expect(screen.getByText(/12 人/)).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "参加" }));
    await waitFor(() => {
      expect(joined).toHaveBeenCalledWith(expect.objectContaining({ channelId: "d" }));
    });

    await userEvent.type(screen.getByRole("searchbox", { name: "チャンネルを検索" }), "相談");
    expect(screen.queryByRole("link", { name: /general/ })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /design/ })).toBeInTheDocument();
  });
});
