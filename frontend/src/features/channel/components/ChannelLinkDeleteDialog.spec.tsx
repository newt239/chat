import { create } from "@bufbuild/protobuf";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelLinkSchema, ChannelLinkService } from "#/gen/chat/v1/channel_link_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelLinkDeleteDialog } from "./ChannelLinkDeleteDialog";

import type { DeleteChannelLinkRequest } from "#/gen/chat/v1/channel_link_service_pb";

describe("ChannelLinkDeleteDialog", () => {
  test("確定するとリンクを削除して onDeleted を呼ぶ", async () => {
    const remove = vi.fn<(req: DeleteChannelLinkRequest) => void>();
    const onDeleted = vi.fn<() => void>();
    await renderWithProviders(
      <ChannelLinkDeleteDialog
        channelId="c1"
        link={create(ChannelLinkSchema, { id: "l1", title: "Figma" })}
        onClose={vi.fn<() => void>()}
        onDeleted={onDeleted}
      />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelLinkService.method.deleteChannelLink, (req) => {
          remove(req);
          return {};
        });
      },
    );
    const dialog = await screen.findByRole("alertdialog", { name: "Figma を削除しますか？" });
    await userEvent.click(within(dialog).getByRole("button", { name: "削除" }));
    await waitFor(() => {
      expect(onDeleted).toHaveBeenCalled();
    });
    expect(remove).toHaveBeenCalledWith(expect.objectContaining({ linkId: "l1" }));
  });
});
