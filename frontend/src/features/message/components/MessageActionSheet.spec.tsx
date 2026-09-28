import { create } from "@bufbuild/protobuf";
import { IconExternalLink, IconTrash } from "@tabler/icons-react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { UserSummarySchema } from "#/gen/chat/v1/user_pb";

import { MessageActionSheet } from "./MessageActionSheet";

const message = create(MessageSchema, {
  body: "デプロイしました",
  id: "m1",
  user: create(UserSummarySchema, { displayName: "Alice", id: "u1" }),
});

const renderSheet = () => {
  const handlers = {
    handleDelete: vi.fn<() => void>(),
    handleOpenChange: vi.fn<(isOpen: boolean) => void>(),
    handleReact: vi.fn<(emoji: string) => void>(),
  };
  render(
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
  );
  return handlers;
};

describe("MessageActionSheet", () => {
  test("メッセージの抜粋と操作を並べ、新しいタブで開く操作は出さない", () => {
    renderSheet();

    const sheet = screen.getByRole("dialog", { name: "メッセージの操作" });
    expect(sheet).toHaveTextContent("Alice");
    expect(sheet).toHaveTextContent("デプロイしました");
    expect(screen.getByRole("menuitem", { name: "メッセージを削除" })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: "スレッドを新しいタブで開く" })).toBeNull();
  });

  test("リアクションや操作を選ぶとシートを閉じる", async () => {
    const handlers = renderSheet();

    await userEvent.click(screen.getByRole("button", { name: "🎉 でリアクション" }));
    expect(handlers.handleReact).toHaveBeenCalledWith("🎉");
    expect(handlers.handleOpenChange).toHaveBeenLastCalledWith(false);

    await userEvent.click(screen.getByRole("menuitem", { name: "メッセージを削除" }));
    expect(handlers.handleDelete).toHaveBeenCalledOnce();
  });
});
