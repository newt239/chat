import { create } from "@bufbuild/protobuf";
import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { CustomEmojiService } from "#/gen/chat/v1/custom_emoji_service_pb";
import { UserSummarySchema } from "#/gen/chat/v1/user_pb";
import { QueryWrapper } from "#/test/QueryWrapper";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ReactionButton } from "./ReactionButton";

const user = (name: string) => create(UserSummarySchema, { displayName: name, id: name });

describe("ReactionButton", () => {
  test("件数と付けた人を読み上げ、自分が付けたかを押下状態で表す", async () => {
    const onPress = vi.fn<() => void>();
    render(
      <ReactionButton
        group={{ count: 2, emoji: "👍", hasUserReacted: true, users: [user("Alice"), user("Bob")] }}
        onPress={onPress}
        onOpenList={vi.fn<() => void>()}
      />,
      { wrapper: QueryWrapper },
    );

    const button = screen.getByRole("button", { name: "👍 2 件。Alice、Bob" });
    expect(button).toHaveAttribute("aria-pressed", "true");
    expect(button).toHaveTextContent("2");

    await userEvent.click(button);
    expect(onPress).toHaveBeenCalledOnce();
  });

  test("5 人以上は先頭の 3 人と残りの人数にまとめ、右クリックで一覧を開く", () => {
    const onOpenList = vi.fn<() => void>();
    const users = ["A", "B", "C", "D", "E"].map((name) => user(name));
    render(
      <ReactionButton
        group={{ count: 5, emoji: "🎉", hasUserReacted: false, users }}
        onPress={vi.fn<() => void>()}
        onOpenList={onOpenList}
      />,
      { wrapper: QueryWrapper },
    );

    const button = screen.getByRole("button", { name: "🎉 5 件。A、B、C ほか 2 人" });
    fireEvent.contextMenu(button);
    expect(onOpenList).toHaveBeenCalledOnce();
  });

  test(":name: のリアクションはカスタム絵文字の画像で出す", async () => {
    await renderWithProviders(
      <ReactionButton
        group={{ count: 1, emoji: ":party:", hasUserReacted: false, users: [user("Alice")] }}
        onPress={vi.fn<() => void>()}
        onOpenList={vi.fn<() => void>()}
      />,
      "/app/ws1",
      (routes) => {
        routes.rpc(CustomEmojiService.method.listCustomEmojis, () => ({
          emojis: [{ id: "e1", imageUrl: "https://example.com/party.png", name: "party" }],
        }));
      },
    );
    const button = screen.getByRole("button", { name: ":party: 1 件。Alice" });
    expect(await within(button).findByRole("img", { name: ":party:" })).toHaveAttribute(
      "src",
      "https://example.com/party.png",
    );
  });
});
