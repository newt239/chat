import { Code, ConnectError } from "@connectrpc/connect";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { LinkService } from "#/gen/chat/v1/link_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { LinkPreviewCard } from "./LinkPreviewCard";

describe("LinkPreviewCard", () => {
  test("URL の OGP を取得してカードに出す", async () => {
    await renderWithProviders(
      <LinkPreviewCard url="https://example.com/post" onRemove={() => {}} />,
      "/app/ws1",
      (routes) => {
        routes.rpc(LinkService.method.fetchOgp, ({ url }) => ({
          ogp: { siteName: "Example", title: `記事 ${url}` },
        }));
      },
    );
    expect(
      await screen.findByRole("link", { name: "記事 https://example.com/post" }),
    ).toBeInTheDocument();
  });

  test("取得できなければその旨と URL を出す", async () => {
    await renderWithProviders(
      <LinkPreviewCard url="https://example.com/post" onRemove={() => {}} />,
      "/app/ws1",
      (routes) => {
        routes.rpc(LinkService.method.fetchOgp, () => {
          throw new ConnectError("failed", Code.Unavailable);
        });
      },
    );
    expect(await screen.findByText(/プレビューを読み込めませんでした/)).toHaveTextContent(
      "https://example.com/post",
    );
  });
});
