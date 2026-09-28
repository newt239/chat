import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelLinkSchema, ChannelLinkService } from "#/gen/chat/v1/channel_link_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelLinksSection } from "./ChannelLinksSection";

import type {
  CreateChannelLinkRequest,
  ReorderChannelLinksRequest,
} from "#/gen/chat/v1/channel_link_service_pb";

const links = [
  create(ChannelLinkSchema, { id: "l1", title: "設計書", url: "https://example.com/spec" }),
  create(ChannelLinkSchema, { id: "l2", title: "Figma", url: "https://figma.com/x" }),
];

const setup = async (canEdit: boolean) => {
  const reorder = vi.fn<(req: ReorderChannelLinksRequest) => void>();
  const createLink = vi.fn<(req: CreateChannelLinkRequest) => void>();
  await renderWithProviders(<ChannelLinksSection channelId="c1" />, "/app/ws1", (routes) => {
    routes.rpc(ChannelLinkService.method.listChannelLinks, () => ({ canEdit, links }));
    routes.rpc(ChannelLinkService.method.reorderChannelLinks, (req) => {
      reorder(req);
      return {};
    });
    routes.rpc(ChannelLinkService.method.createChannelLink, (req) => {
      createLink(req);
      return {};
    });
  });
  await screen.findByRole("link", { name: "設計書" });
  return { createLink, reorder };
};

describe("ChannelLinksSection", () => {
  test("下へ移動すると全リンクの新しい並びを送る", async () => {
    const { reorder } = await setup(true);
    await userEvent.click(screen.getByRole("button", { name: "設計書 を下へ" }));
    await waitFor(() => {
      expect(reorder).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", linkIds: ["l2", "l1"] }),
      );
    });
  });

  test("表示名を空にして追加するとドメイン名を表示名にする", async () => {
    const { createLink } = await setup(true);
    await userEvent.click(screen.getByRole("button", { name: "追加" }));
    const url = await screen.findByRole("textbox", { name: "URL" });
    await userEvent.clear(url);
    await userEvent.type(url, "https://docs.example.com/a");
    await userEvent.click(screen.getAllByRole("button", { name: "追加" }).at(-1) ?? url);
    await waitFor(() => {
      expect(createLink).toHaveBeenCalledWith(
        expect.objectContaining({
          channelId: "c1",
          title: "docs.example.com",
          url: "https://docs.example.com/a",
        }),
      );
    });
  });

  test("URL が不正なら追加しない", async () => {
    const { createLink } = await setup(true);
    await userEvent.click(screen.getByRole("button", { name: "追加" }));
    const url = await screen.findByRole("textbox", { name: "URL" });
    await userEvent.clear(url);
    await userEvent.type(url, "not a url");
    await userEvent.click(screen.getAllByRole("button", { name: "追加" }).at(-1) ?? url);
    expect(
      await screen.findByText("https:// から始まる URL を入力してください"),
    ).toBeInTheDocument();
    expect(createLink).not.toHaveBeenCalled();
  });

  test("編集できない人には操作を出さない", async () => {
    await setup(false);
    expect(screen.queryByRole("button", { name: "追加" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "設計書 を編集" })).not.toBeInTheDocument();
  });
});
