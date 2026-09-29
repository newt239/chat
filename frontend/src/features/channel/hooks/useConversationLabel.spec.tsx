import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import {
  DirectMessageSchema,
  DirectMessageService,
  DirectMessageType,
} from "#/gen/chat/v1/direct_message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { useConversationLabel } from "./useConversationLabel";

const Labels = () => {
  const labelOf = useConversationLabel("ws1");
  return <output>{["c1", "d1", "unknown"].map((id) => `[${labelOf(id)}]`).join("")}</output>;
};

describe("useConversationLabel", () => {
  test("チャンネルは #名前、DM は相手の名前、分からなければ空文字にする", async () => {
    await renderWithProviders(<Labels />, "/app/ws1", (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [create(ChannelSchema, { id: "c1", name: "general" })],
      }));
      routes.rpc(DirectMessageService.method.listDirectMessages, () => ({
        directMessages: [
          create(DirectMessageSchema, {
            id: "d1",
            members: [{ displayName: "Bob", userId: "u-bob" }],
            type: DirectMessageType.DM,
          }),
        ],
      }));
    });

    expect(await screen.findByText("[#general][Bob][]")).toBeInTheDocument();
  });
});
