import { create } from "@bufbuild/protobuf";
import { act, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { mediaPlayer } from "../mediaPlayer";
import { MiniPlayer } from "./MiniPlayer";

const render = (variant: "sidebar" | "mobile") =>
  renderWithProviders(<MiniPlayer variant={variant} />, "/app/ws1/other", (routes) => {
    routes.rpc(ChannelService.method.listChannels, () => ({
      channels: [create(ChannelSchema, { id: "ch1", name: "dev/general" })],
    }));
  });

const playAudio = () =>
  act(() =>
    mediaPlayer.play(
      {
        attachmentId: "a1",
        authorName: "Alice",
        channelId: "ch1",
        durationSeconds: 90,
        fileName: "standup.m4a",
        kind: "audio",
        messageId: "m1",
        parentId: undefined,
        workspaceId: "ws1",
      },
      () => Promise.resolve("https://storage.example.com/a1"),
    ),
  );

beforeEach(() => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockReturnValue();
  vi.spyOn(HTMLMediaElement.prototype, "load").mockReturnValue();
});

afterEach(() => {
  act(() => {
    mediaPlayer.stop();
  });
  vi.restoreAllMocks();
});

describe("MiniPlayer", () => {
  test("再生していなければ何も出さない", async () => {
    await render("sidebar");
    expect(screen.queryByRole("region")).toBeNull();
  });

  test("再生中のメッセージが画面にないときに出し、元のメッセージへ移動できる", async () => {
    const { router } = await render("sidebar");
    await playAudio();

    const player = await screen.findByRole("region", { name: "再生中のメディア" });
    expect(player).toHaveTextContent("standup.m4a");
    expect(await screen.findByText("#general · Alice")).toBeInTheDocument();
    expect(screen.getByRole("slider", { name: "再生位置" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /元のメッセージへ移動/ }));
    await waitFor(() => {
      expect(router.state.location.href).toBe("/app/ws1/ch1?message=m1");
    });

    await userEvent.click(screen.getByRole("button", { name: "再生を終了" }));
    expect(mediaPlayer.getState().track).toBeNull();
    await waitFor(() => {
      expect(screen.queryByRole("region")).toBeNull();
    });
  });

  test("画面幅に合わない配置では描画しない", async () => {
    await render("mobile");
    await playAudio();
    expect(screen.queryByRole("region")).toBeNull();
  });
});
