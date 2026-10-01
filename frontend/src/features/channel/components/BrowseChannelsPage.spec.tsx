import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";
import {
  BrowsableChannelMembership,
  BrowsableChannelSchema,
  BrowsableChannelSort,
  ChannelSchema,
  ChannelService,
} from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { BrowseChannelsPage } from "./BrowseChannelsPage";

import type { JoinChannelRequest } from "#/gen/chat/v1/channel_member_service_pb";
import type { SearchBrowsableChannelsRequest } from "#/gen/chat/v1/channel_service_pb";

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
  test("未参加のチャンネルに参加でき、検索・絞り込み・並び順・ページを条件にして取得する", async () => {
    const joined = vi.fn<(req: JoinChannelRequest) => void>();
    const searched = vi.fn<(req: SearchBrowsableChannelsRequest) => void>();
    await renderWithProviders(<BrowseChannelsPage />, "/app/ws1/browse-channels", (routes) => {
      routes.rpc(ChannelService.method.searchBrowsableChannels, (req) => {
        searched(req);
        return { channels, total: 45 };
      });
      routes.rpc(ChannelMemberService.method.joinChannel, (req) => {
        joined(req);
        return {};
      });
    });

    expect(await screen.findByRole("link", { name: /design/ })).toBeInTheDocument();
    expect(screen.getByText("参加中", { selector: "span" })).toBeInTheDocument();
    expect(screen.getByText(/12 人/)).toBeInTheDocument();
    expect(screen.getByText("45 件")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "参加" }));
    await waitFor(() => {
      expect(joined).toHaveBeenCalledWith(expect.objectContaining({ channelId: "d" }));
    });

    await userEvent.click(screen.getByRole("button", { name: "次のページ" }));
    await waitFor(() => {
      expect(searched).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2, perPage: 20 }));
    });
    expect(screen.getByText("2 / 3 ページ")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("radio", { name: "未参加" }));
    await userEvent.type(screen.getByRole("searchbox", { name: "チャンネルを検索" }), "相談");
    await waitFor(() => {
      expect(searched).toHaveBeenLastCalledWith(
        expect.objectContaining({
          membership: BrowsableChannelMembership.NOT_JOINED,
          page: 1,
          query: "相談",
          sort: BrowsableChannelSort.UNSPECIFIED,
        }),
      );
    });
  });
});
