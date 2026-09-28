import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelChip } from "./ChannelChip";

describe("ChannelChip", () => {
  test("一覧にない未参加のチャンネルは個別に取得し、親からの相対パスで出す", async () => {
    await renderWithProviders(
      <ChannelChip workspaceId="ws1" parentName="dev" channelId="be" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelService.method.listChannels, () => ({
          channels: [create(ChannelSchema, { id: "dev", name: "dev" })],
        }));
        routes.rpc(ChannelService.method.getChannel, () => ({
          channel: create(ChannelSchema, { id: "be", name: "dev/backend/api" }),
        }));
      },
    );

    const chip = await screen.findByRole("link", { name: "#backend/api を開く" });
    expect(chip).toHaveTextContent("# backend/api");
    expect(chip).toHaveAttribute("href", "/app/ws1/be");
  });
});
