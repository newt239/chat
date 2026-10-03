import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageSchema, ReactionSchema } from "#/gen/chat/v1/message_pb";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { ReactionsDialog } from "./ReactionsDialog";

const reaction = (emoji: string, user: { id: string; displayName: string }, minute: number) =>
  create(ReactionSchema, {
    createdAt: timestampFromDate(new Date(2026, 8, 28, 10, minute)),
    emoji,
    user,
  });

const message = create(MessageSchema, {
  body: "リリースしました",
  id: "m1",
  reactions: [
    reaction("🎉", { displayName: "Bob", id: "u2" }, 1),
    reaction("🎉", currentUser, 3),
    reaction("👀", { displayName: "Carol", id: "u3" }, 2),
  ],
  user: { displayName: "Bob", id: "u2" },
});

describe("ReactionsDialog", () => {
  test("すべてのタブでは新しい順に並べ、絵文字のタブで絞り込み、自分のものは取り消せる", async () => {
    const onToggleReaction = vi.fn<(emoji: string) => void>();
    const onTabChange = vi.fn<(tab: string | null) => void>();
    await renderWithProviders(
      <ReactionsDialog
        message={message}
        tab="all"
        onTabChange={onTabChange}
        onToggleReaction={onToggleReaction}
      />,
      "/app/ws1/ch1",
      () => {},
    );

    const dialog = screen.getByRole("dialog", { name: "リアクション" });
    expect(within(dialog).getByText("Bob: リリースしました")).toBeInTheDocument();
    const rows = within(dialog).getAllByRole("listitem");
    expect(rows.map((row) => row.textContent)).toEqual([
      expect.stringContaining("あなた"),
      expect.stringContaining("Carol"),
      expect.stringContaining("Bob"),
    ]);

    await userEvent.click(within(dialog).getByRole("tab", { name: /👀/ }));
    expect(onTabChange).toHaveBeenCalledWith("👀");

    await userEvent.click(within(dialog).getByRole("button", { name: "取り消す" }));
    expect(onToggleReaction).toHaveBeenCalledWith("🎉");
  });

  test("選んだ絵文字のタブを開く", async () => {
    await renderWithProviders(
      <ReactionsDialog
        message={message}
        tab="👀"
        onTabChange={vi.fn<(tab: string | null) => void>()}
        onToggleReaction={vi.fn<(emoji: string) => void>()}
      />,
      "/app/ws1/ch1",
      () => {},
    );

    const rows = within(screen.getByRole("tabpanel")).getAllByRole("listitem");
    expect(rows).toHaveLength(1);
    expect(rows[0]).toHaveTextContent("Carol");
  });
});
