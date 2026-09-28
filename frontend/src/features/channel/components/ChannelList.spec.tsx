import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelList } from "./ChannelList";

const channels = [
  create(ChannelSchema, { id: "dev", isMember: true, name: "dev" }),
  create(ChannelSchema, {
    hasMention: true,
    id: "fe",
    isMember: true,
    name: "dev/frontend",
    parentId: "dev",
    unreadCount: 2,
  }),
  create(ChannelSchema, {
    id: "web",
    isMember: true,
    name: "dev/frontend/web",
    parentId: "fe",
    unreadCount: 3,
  }),
  create(ChannelSchema, { id: "g", isMember: true, name: "general" }),
];

const render = () =>
  renderWithProviders(<ChannelList workspaceId="ws1" />, "/app/ws1", (routes) => {
    routes.rpc(ChannelService.method.listChannels, () => ({ channels }));
  });

describe("ChannelList", () => {
  test("階層をツリーで並べ、行には末尾の名前だけを出す", async () => {
    await render();
    expect(await screen.findByRole("link", { name: /^公開チャンネル\s?dev$/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /^公開チャンネル\s?frontend/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /^公開チャンネル\s?web$/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /^公開チャンネル\s?general$/ })).toBeInTheDocument();
  });

  test("折りたたむと子の数と子孫の未読の合計を親に出す", async () => {
    await render();
    await userEvent.click(await screen.findByRole("button", { name: "dev の下階層を折りたたむ" }));

    const parent = await screen.findByRole("link", { name: /^公開チャンネル\s?dev/ });
    expect(parent).toHaveTextContent("1");
    expect(parent).toHaveTextContent("5");
    // 折りたたみのアニメーションが終わってから消える
    await waitFor(() => {
      expect(screen.queryByRole("link", { name: /frontend/ })).not.toBeInTheDocument();
    });

    await userEvent.click(screen.getByRole("button", { name: "dev の下階層を展開" }));
    expect(await screen.findByRole("link", { name: /frontend/ })).toBeInTheDocument();
  });
});
