import { create } from "@bufbuild/protobuf";
import { IconExternalLink, IconTrash } from "@tabler/icons-react";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { UserSummarySchema } from "#/gen/chat/v1/user_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MessageActionSheet } from "./MessageActionSheet";

const message = create(MessageSchema, {
  body: "デプロイしました",
  id: "m1",
  user: create(UserSummarySchema, { displayName: "Alice", id: "u1" }),
});

const renderSheet = async () => {
  const handlers = {
    handleDelete: vi.fn<() => void>(),
    handleOpenChange: vi.fn<(isOpen: boolean) => void>(),
    handleReact: vi.fn<(emoji: string) => void>(),
  };
  await renderWithProviders(
    <MessageActionSheet
      isOpen
      onOpenChange={handlers.handleOpenChange}
      message={message}
      actions={[
        {
          href: "/app/w/c?message=m1",
          icon: IconExternalLink,
          id: "threadInNewTab",
          label: "スレッドを新しいタブで開く",
          tone: "default",
        },
        {
          icon: IconTrash,
          id: "delete",
          label: "メッセージを削除",
          onAction: handlers.handleDelete,
          tone: "danger",
        },
      ]}
      onReact={handlers.handleReact}
    />,
    "/app/ws1",
    () => {},
  );
  return handlers;
};

describe("MessageActionSheet", () => {
  test("メッセージの抜粋と操作を並べ、新しいタブで開く操作は出さない", async () => {
    await renderSheet();

    const sheet = await screen.findByRole("dialog", { name: "メッセージの操作" });
    expect(sheet).toHaveTextContent("Alice");
    expect(sheet).toHaveTextContent("デプロイしました");
    expect(screen.getByRole("menuitem", { name: "メッセージを削除" })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: "スレッドを新しいタブで開く" })).toBeNull();
  });

  test("リアクションや操作を選ぶとシートを閉じる", async () => {
    const handlers = await renderSheet();

    await userEvent.click(await screen.findByRole("button", { name: "🎉 でリアクション" }));
    expect(handlers.handleReact).toHaveBeenCalledWith("🎉");
    expect(handlers.handleOpenChange).toHaveBeenLastCalledWith(false);

    await userEvent.click(screen.getByRole("menuitem", { name: "メッセージを削除" }));
    expect(handlers.handleDelete).toHaveBeenCalledOnce();
  });
});
