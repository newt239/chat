import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelLinkSchema, ChannelLinkService } from "#/gen/chat/v1/channel_link_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelLinkDialogLoader } from "./ChannelLinkDialogLoader";

const render = (linkId: string | null, canEdit: boolean) =>
  renderWithProviders(
    <ChannelLinkDialogLoader channelId="c1" linkId={linkId} onClose={vi.fn<() => void>()} />,
    "/app/ws1/c1",
    (routes) => {
      routes.rpc(ChannelLinkService.method.listChannelLinks, () => ({
        canEdit,
        links: [create(ChannelLinkSchema, { id: "l1", title: "設計書", url: "https://a.com" })],
      }));
    },
  );

describe("ChannelLinkDialogLoader", () => {
  test("ID のリンクを読み込んでから編集のダイアログを開く", async () => {
    await render("l1", true);
    expect(await screen.findByRole("textbox", { name: "表示名" })).toHaveValue("設計書");
    expect(screen.getByRole("textbox", { name: "URL" })).toHaveValue("https://a.com");
  });

  test("見つからないリンクや編集できない人には開かない", async () => {
    await render("missing", true);
    await render(null, false);
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).toBeNull();
    });
  });
});
