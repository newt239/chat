import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ChannelService } from "#/gen/chat/v1/channel_service_pb";
import {
  DirectMessageSchema,
  DirectMessageService,
  DirectMessageType,
} from "#/gen/chat/v1/direct_message_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MessageListCard } from "./MessageListCard";

const render = (parentId?: string) =>
  renderWithProviders(
    <MessageListCard
      workspaceId="ws1"
      message={create(MessageSchema, { channelId: "d1", id: "m1", parentId })}
    >
      本文
    </MessageListCard>,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({ channels: [] }));
      routes.rpc(DirectMessageService.method.listDirectMessages, () => ({
        directMessages: [
          create(DirectMessageSchema, {
            id: "d1",
            members: [{ displayName: "Bob", userId: "u-bob" }],
            type: DirectMessageType.DM,
          }),
        ],
      }));
    },
  );

describe("MessageListCard", () => {
  test("会話の名前と、チャンネルでの位置へのリンクを出す", async () => {
    await render();
    expect(await screen.findByText("Bob")).toBeInTheDocument();
    expect(screen.getByText("本文")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "チャンネルで表示" })).toHaveAttribute(
      "href",
      "/app/ws1/d1?message=m1",
    );
  });

  test("返信はスレッドでの位置へリンクする", async () => {
    await render("p1");
    expect(await screen.findByText(/スレッド内/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "スレッドで表示" })).toHaveAttribute(
      "href",
      "/app/ws1/d1/thread/p1?message=m1",
    );
  });
});
