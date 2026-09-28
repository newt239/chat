import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { mediaPlayer } from "#/features/player/mediaPlayer";
import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";
import { MessageAttachmentSchema, MessageSchema } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AudioAttachment } from "./AudioAttachment";

beforeEach(() => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockReturnValue();
  vi.spyOn(HTMLMediaElement.prototype, "load").mockReturnValue();
});

afterEach(() => {
  mediaPlayer.stop();
  vi.restoreAllMocks();
});

describe("AudioAttachment", () => {
  test("再生ボタンで署名付き URL を取得し、アプリ全体のプレイヤーで再生する", async () => {
    const getDownloadUrl = vi.fn<() => { url: string }>(() => ({
      url: "https://storage.example.com/voice",
    }));
    await renderWithProviders(
      <AudioAttachment
        attachment={create(MessageAttachmentSchema, {
          fileName: "voice.m4a",
          id: "a1",
          media: { durationSeconds: 83 },
          mimeType: "audio/mp4",
        })}
        message={create(MessageSchema, { channelId: "ch1", id: "m1" })}
      />,
      "/app/ws1/ch1",
      (routes) => {
        routes.rpc(AttachmentService.method.getDownloadUrl, getDownloadUrl);
      },
    );

    expect(screen.getByText("voice.m4a")).toBeInTheDocument();
    expect(screen.getByText("0:00 / 1:23")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "再生" }));
    await waitFor(() => {
      expect(mediaPlayer.getState()).toMatchObject({
        isPlaying: true,
        track: { attachmentId: "a1", channelId: "ch1", workspaceId: "ws1" },
      });
    });
    expect(getDownloadUrl).toHaveBeenCalledOnce();
    expect(screen.getByRole("button", { name: "一時停止" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "再生速度" }));
    expect(screen.getByText("1.5x")).toBeInTheDocument();
  });
});
