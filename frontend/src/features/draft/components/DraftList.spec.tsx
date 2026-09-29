import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DraftSchema, DraftService } from "#/gen/chat/v1/draft_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { DraftList } from "./DraftList";

import type { DeleteDraftRequest, Draft } from "#/gen/chat/v1/draft_service_pb";

const render = (drafts: Draft[], deleteDraft: (req: DeleteDraftRequest) => void) =>
  renderWithProviders(<DraftList workspaceId="ws1" />, "/app/ws1/drafts", (routes) => {
    routes.rpc(DraftService.method.listDrafts, () => ({ drafts }));
    routes.rpc(DraftService.method.deleteDraft, (req) => {
      deleteDraft(req);
      return {};
    });
    routes.rpc(ChannelService.method.listChannels, () => ({
      channels: [create(ChannelSchema, { id: "c1", name: "general" })],
    }));
  });

describe("DraftList", () => {
  test("下書きを並べ、削除できる", async () => {
    const deleteDraft = vi.fn<(req: DeleteDraftRequest) => void>();
    await render(
      [create(DraftSchema, { body: "続きを書く", channelId: "c1", id: "d1", parentId: "m1" })],
      deleteDraft,
    );

    expect(await screen.findByText("続きを書く")).toBeInTheDocument();
    expect(await screen.findByText("#general")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "下書きを削除" }));
    await waitFor(() => {
      expect(deleteDraft).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", parentId: "m1" }),
      );
    });
  });

  test("下書きがなければ案内を出す", async () => {
    await render([], vi.fn<(req: DeleteDraftRequest) => void>());
    expect(await screen.findByText("下書きはありません")).toBeInTheDocument();
  });
});
