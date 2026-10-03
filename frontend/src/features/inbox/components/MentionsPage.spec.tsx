import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { MentionCursorSchema, MentionService } from "#/gen/chat/v1/mention_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MentionsPage } from "./MentionsPage";

import type { ListMentionsRequest } from "#/gen/chat/v1/mention_service_pb";
import type { CreateMessageRequest } from "#/gen/chat/v1/message_service_pb";

// 末尾が見えたときの読み込みをテストから起こせるようにする
const observers: ((entries: { isIntersecting: boolean }[]) => void)[] = [];
vi.stubGlobal(
  "IntersectionObserver",
  class {
    public constructor(callback: (entries: { isIntersecting: boolean }[]) => void) {
      observers.push(callback);
    }
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

const setup = async () => {
  const listed = vi.fn<(req: ListMentionsRequest) => void>();
  const created = vi.fn<(req: CreateMessageRequest) => void>();
  await renderWithProviders(<MentionsPage />, "/app/ws1", (routes) => {
    routes.rpc(ChannelService.method.listChannels, () => ({
      channels: [create(ChannelSchema, { id: "c1", name: "design" })],
    }));
    routes.rpc(DirectMessageService.method.listDirectMessages, () => ({ directMessages: [] }));
    routes.rpc(MentionService.method.listMentions, (req) => {
      listed(req);
      return req.cursor === undefined
        ? {
            messages: [message("m1", "確認お願いします"), message("m2", "返信のメンション", "p1")],
            nextCursor: create(MentionCursorSchema, { messageId: "m2" }),
          }
        : { messages: [message("m3", "古いメンション")] };
    });
    routes.rpc(MessageService.method.createMessage, (req) => {
      created(req);
      return { message: message("r1", req.body, req.parentId) };
    });
  });
  return { created, listed };
};

describe("MentionsPage", () => {
  test("メンションを並べ、末尾が見えたら次のページを読み込む", async () => {
    const { listed } = await setup();
    expect(await screen.findByText("確認お願いします")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "スレッドで表示" })).toHaveAttribute(
      "href",
      "/app/ws1/c1/thread/p1?message=m2",
    );

    await waitFor(() => {
      expect(observers.length).toBeGreaterThan(0);
    });
    observers.at(-1)?.([{ isIntersecting: true }]);
    expect(await screen.findByText("古いメンション")).toBeInTheDocument();
    expect(listed.mock.lastCall?.[0].cursor?.messageId).toBe("m2");
  });

  test("一覧からスレッドに返信し、送った返信をカードに出す", async () => {
    const { created } = await setup();
    const [first, second] = await screen.findAllByRole("textbox", {
      name: "Bob さんにスレッドで返信…",
    });
    await userEvent.type(first ?? document.body, "見ます{Enter}");
    await waitFor(() => {
      expect(created).toHaveBeenCalledWith(
        expect.objectContaining({ body: "見ます", channelId: "c1", parentId: "m1" }),
      );
    });
    expect(await screen.findByText("見ます")).toBeInTheDocument();

    // 返信へのメンションには同じスレッドで返す
    await userEvent.type(second ?? document.body, "了解{Enter}");
    await waitFor(() => {
      expect(created).toHaveBeenLastCalledWith(expect.objectContaining({ parentId: "p1" }));
    });
  });

  test("メンションがなければ案内を出す", async () => {
    await renderWithProviders(<MentionsPage />, "/app/ws1", (routes) => {
      routes.rpc(MentionService.method.listMentions, () => ({ messages: [] }));
    });
    expect(await screen.findByText("メンションはありません")).toBeInTheDocument();
  });
});
