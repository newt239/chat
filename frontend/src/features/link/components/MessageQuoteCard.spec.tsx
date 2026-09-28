import { create } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { screen, waitFor } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageLinkSchema, MessagePreviewSchema } from "#/gen/chat/v1/message_pb";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MessageQuoteCard } from "./MessageQuoteCard";

import type { MessageLink } from "#/gen/chat/v1/message_pb";

const preview = create(MessagePreviewSchema, {
  bodyExcerpt: "**リリース**は金曜です",
  channelId: "ch2",
  channelName: "dev/release",
  messageId: "m2",
  user: { displayName: "Bob", id: "u2" },
});

const render = (link: MessageLink, getMessagePreview: () => { preview: typeof preview }) =>
  renderWithProviders(<MessageQuoteCard link={link} />, "/app/ws1/ch1", (routes) => {
    routes.rpc(MessageService.method.getMessagePreview, getMessagePreview);
  });

describe("MessageQuoteCard", () => {
  test("投稿者・チャンネル・抜粋を出し、元のメッセージへのリンクを置く", async () => {
    const getMessagePreview = vi.fn<() => { preview: typeof preview }>();
    await render(
      create(MessageLinkSchema, { linkedMessageId: "m2", messagePreview: preview }),
      getMessagePreview,
    );

    expect(screen.getByText("Bob")).toBeInTheDocument();
    expect(screen.getByText(/#release ·/)).toBeInTheDocument();
    expect(screen.getByText("リリースは金曜です")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "メッセージを表示" })).toHaveAttribute(
      "href",
      "/app/ws1/ch2?message=m2",
    );
    expect(getMessagePreview).not.toHaveBeenCalled();
  });

  test("引用が含まれていなければ取得して表示する", async () => {
    await render(create(MessageLinkSchema, { linkedMessageId: "m2" }), () => ({ preview }));

    expect(await screen.findByText("リリースは金曜です")).toBeInTheDocument();
  });

  test("閲覧できないメッセージは表示しない", async () => {
    const getMessagePreview = vi.fn<() => { preview: typeof preview }>(() => {
      throw new ConnectError("not found", Code.NotFound);
    });
    await render(create(MessageLinkSchema, { linkedMessageId: "m3" }), getMessagePreview);

    await waitFor(() => {
      expect(getMessagePreview).toHaveBeenCalledOnce();
    });
    expect(screen.queryByRole("link")).toBeNull();
  });
});
