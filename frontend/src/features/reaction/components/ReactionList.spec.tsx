import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ReactionSchema } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ReactionList } from "./ReactionList";

const emojis = ["👍", "🎉", "👀", "🙏", "🔥", "💯", "😂", "🚀", "✅", "❤️", "⭐", "☕"];

describe("ReactionList", () => {
  test("10 種類を超える分は +N にまとめ、押すと広げる", async () => {
    await renderWithProviders(
      <ReactionList
        messageId="m1"
        reactions={emojis.map((emoji) => create(ReactionSchema, { emoji, user: { id: "u2" } }))}
        onOpenList={vi.fn<(emoji: string) => void>()}
      />,
      "/app/ws1/ch1",
      () => {},
    );

    expect(screen.getAllByRole("button", { name: /件。/ })).toHaveLength(10);
    await userEvent.click(screen.getByRole("button", { name: "ほかのリアクションを表示" }));
    expect(screen.getAllByRole("button", { name: /件。/ })).toHaveLength(12);
    expect(screen.getByRole("button", { name: "折りたたむ" })).toBeInTheDocument();
  });
});
