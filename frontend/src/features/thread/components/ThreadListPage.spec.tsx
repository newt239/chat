import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { ParticipatingThreadSchema, ThreadService } from "#/gen/chat/v1/thread_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ThreadListPage } from "./ThreadListPage";

import type { CreateMessageRequest } from "#/gen/chat/v1/message_service_pb";

vi.stubGlobal(
  "IntersectionObserver",
  class {
    public observe = () => {};
    public disconnect = () => {};
  },
);

const message = (id: string, body: string, parentId?: string) =>
  create(MessageSchema, {
    body,
    channelId: "c1",
    createdAt: timestampFromDate(new Date(2026, 8, 28, 10, 0)),
    id,
    parentId,
    user: { displayName: "Bob", id: "u-bob" },
    userId: "u-bob",
  });

describe("ThreadListPage", () => {
  test("カードに親と最新の返信を出し、その場で返信すると一覧に足す", async () => {
    const created = vi.fn<(req: CreateMessageRequest) => void>();
    const list = vi.fn();
    await renderWithProviders(<ThreadListPage />, "/app/ws1", (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [create(ChannelSchema, { id: "c1", name: "design" })],
      }));
      routes.rpc(DirectMessageService.method.listDirectMessages, () => ({ directMessages: [] }));
      routes.rpc(ThreadService.method.listParticipatingThreads, () => {
        list();
        return {
          threads: [
            create(ParticipatingThreadSchema, {
              firstMessage: message("p1", "親の投稿"),
              latestReplies: [message("r2", "返信 2", "p1"), message("r3", "返信 3", "p1")],
              replyCount: 3,
              threadId: "p1",
              unreadCount: 1,
            }),
          ],
        };
      });
      routes.rpc(MessageService.method.createMessage, (req) => {
        created(req);
        return { message: message("r4", req.body, "p1") };
      });
    });

    expect(await screen.findByText("親の投稿")).toBeInTheDocument();
    expect(await screen.findByText("#design")).toBeInTheDocument();
    expect(screen.getByText("返信 3")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "他 1 件の返信を表示" })).toHaveAttribute(
      "href",
      "/app/ws1/c1/thread/p1",
    );
    expect(screen.getByText("未読 1")).toBeInTheDocument();

    await userEvent.type(screen.getByRole("textbox", { name: "返信する…" }), "了解です{Enter}");

    await waitFor(() => {
      expect(created).toHaveBeenCalledWith(
        expect.objectContaining({ body: "了解です", channelId: "c1", parentId: "p1" }),
      );
    });
    expect(await screen.findByText("了解です")).toBeInTheDocument();
    expect(screen.queryByText("返信 2")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "他 2 件の返信を表示" })).toBeInTheDocument();
    // 一覧は取り直さず、キャッシュを書き換えて反映する
    expect(list).toHaveBeenCalledTimes(1);
  });

  test("参加中のスレッドがなければ案内を出す", async () => {
    await renderWithProviders(<ThreadListPage />, "/app/ws1", (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({ channels: [] }));
      routes.rpc(ThreadService.method.listParticipatingThreads, () => ({ threads: [] }));
    });
    expect(await screen.findByText("参加中のスレッドはありません")).toBeInTheDocument();
  });
});
