import { IconLink } from "@tabler/icons-react";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { PermissionService } from "#/gen/chat/v1/permission_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MessageToolbar } from "./MessageToolbar";

const renderToolbar = async () => {
  const handlers = {
    handleCopyLink: vi.fn<() => void>(),
    handleOverlayOpenChange: vi.fn<(isOpen: boolean) => void>(),
    handleReact: vi.fn<(emoji: string) => void>(),
    handleReplyInThread: vi.fn<() => void>(),
    handleToggleBookmark: vi.fn<() => void>(),
  };
  await renderWithProviders(
    <MessageToolbar
      actions={[
        {
          icon: IconLink,
          id: "copyLink",
          label: "リンクをコピー",
          onAction: handlers.handleCopyLink,
          tone: "default",
        },
        {
          href: "/app/w/c?message=m",
          icon: IconLink,
          id: "threadInNewTab",
          label: "スレッドを新しいタブで開く",
          tone: "default",
        },
      ]}
      isBookmarked={false}
      onToggleBookmark={handlers.handleToggleBookmark}
      onReplyInThread={handlers.handleReplyInThread}
      onReact={handlers.handleReact}
      onOverlayOpenChange={handlers.handleOverlayOpenChange}
    />,
    "/app/ws1",
    (routes) => {
      routes.rpc(PermissionService.method.getPermissions, () => ({}));
    },
  );
  return handlers;
};

describe("MessageToolbar", () => {
  test("よく使うリアクションとスレッド・ブックマークをボタンで操作できる", async () => {
    const handlers = await renderToolbar();

    expect(screen.getByRole("toolbar", { name: "メッセージ操作" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "👍 でリアクション" }));
    await userEvent.click(screen.getByRole("button", { name: "スレッドで返信" }));
    await userEvent.click(screen.getByRole("button", { name: "ブックマーク" }));

    expect(handlers.handleReact).toHaveBeenCalledWith("👍");
    expect(handlers.handleReplyInThread).toHaveBeenCalledOnce();
    expect(handlers.handleToggleBookmark).toHaveBeenCalledOnce();
  });

  test("その他メニューを開いている間はそれを伝え、新しいタブで開く項目はリンクになる", async () => {
    const handlers = await renderToolbar();

    await userEvent.click(screen.getByRole("button", { name: "その他" }));
    expect(handlers.handleOverlayOpenChange).toHaveBeenLastCalledWith(true);

    const newTab = screen.getByRole("menuitem", { name: /スレッドを新しいタブで開く/ });
    expect(newTab).toHaveAttribute("href", "/app/w/c?message=m");
    expect(newTab).toHaveAttribute("target", "_blank");

    await userEvent.click(screen.getByRole("menuitem", { name: /リンクをコピー/ }));
    expect(handlers.handleCopyLink).toHaveBeenCalledOnce();
    expect(handlers.handleOverlayOpenChange).toHaveBeenLastCalledWith(false);
  });
});
