import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { CustomEmojiService } from "#/gen/chat/v1/custom_emoji_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { renderMarkdown } from "./renderer";

describe("renderMarkdown", () => {
  test("登録済みの :name: は画像にし、未登録や時刻はそのまま出す", async () => {
    await renderWithProviders(
      <p>{renderMarkdown("やった :party: :nope: 10:30:45", [])}</p>,
      "/app/ws1",
      (routes) => {
        routes.rpc(CustomEmojiService.method.listCustomEmojis, () => ({
          emojis: [{ id: "e1", imageUrl: "https://example.com/party.png", name: "party" }],
        }));
      },
    );
    expect(await screen.findByRole("img", { name: ":party:" })).toHaveAttribute(
      "src",
      "https://example.com/party.png",
    );
    expect(screen.getByText(/:nope:/)).toBeInTheDocument();
    expect(screen.getByText(/10:30:45/)).toBeInTheDocument();
  });

  test("hiddenUrls のリンクは本文から外し、空になった段落も出さない", async () => {
    const hidden = "https://chat.localhost/app/ws1/ch1?message=m1";
    await renderWithProviders(
      <div>
        {renderMarkdown(`${hidden}\n\nこちらも https://example.com と ${hidden}`, [hidden])}
      </div>,
      "/app/ws1",
      () => {},
    );
    expect(await screen.findByRole("link", { name: "https://example.com" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: hidden })).not.toBeInTheDocument();
    expect(document.body.querySelectorAll("p")).toHaveLength(1);
  });
});
