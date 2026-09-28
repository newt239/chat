import { create } from "@bufbuild/protobuf";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelLinkSchema, ChannelLinkService } from "#/gen/chat/v1/channel_link_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelLinkBar } from "./ChannelLinkBar";

import type { UpdateChannelLinkRequest } from "#/gen/chat/v1/channel_link_service_pb";

describe("ChannelLinkBar", () => {
  test("右クリックの「編集」でリンクを編集する", async () => {
    const update = vi.fn<(req: UpdateChannelLinkRequest) => void>();
    await renderWithProviders(<ChannelLinkBar channelId="c1" />, "/app/ws1", (routes) => {
      routes.rpc(ChannelLinkService.method.listChannelLinks, () => ({
        canEdit: true,
        links: [create(ChannelLinkSchema, { id: "l1", title: "設計書", url: "https://a.com" })],
      }));
      routes.rpc(ChannelLinkService.method.updateChannelLink, (req) => {
        update(req);
        return {};
      });
    });

    fireEvent.contextMenu(await screen.findByRole("link", { name: "設計書" }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "編集" }));
    const title = await screen.findByRole("textbox", { name: "表示名" });
    await userEvent.clear(title);
    await userEvent.type(title, "仕様書");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => {
      expect(update).toHaveBeenCalledWith(
        expect.objectContaining({ linkId: "l1", title: "仕様書", url: "https://a.com" }),
      );
    });
  });

  test("リンクがなければ何も出さない", async () => {
    await renderWithProviders(<ChannelLinkBar channelId="c1" />, "/app/ws1", (routes) => {
      routes.rpc(ChannelLinkService.method.listChannelLinks, () => ({ canEdit: true, links: [] }));
    });
    await waitFor(() => {
      expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
    });
  });
});
