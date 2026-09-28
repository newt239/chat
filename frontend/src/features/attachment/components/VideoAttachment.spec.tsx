import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { mediaPlayer } from "#/features/player/mediaPlayer";
import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";
import { MessageAttachmentSchema, MessageSchema } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { VideoAttachment } from "./VideoAttachment";

beforeEach(() => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockReturnValue();
  vi.spyOn(HTMLMediaElement.prototype, "load").mockReturnValue();
});

afterEach(() => {
  mediaPlayer.stop();
  vi.restoreAllMocks();
});

const render = async (thumbnail: { width: number; height: number } | undefined) => {
  const getDownloadUrl = vi.fn(
    ({ attachmentId, thumbnail: isThumbnail }: { attachmentId: string; thumbnail: boolean }) => ({
      url: `https://storage.example.com/${attachmentId}${isThumbnail ? "-thumbnail" : ""}`,
    }),
  );
  await renderWithProviders(
    <VideoAttachment
      attachment={create(MessageAttachmentSchema, {
        fileName: "demo.mp4",
        id: "v1",
        media: { height: 720, thumbnail, width: 1280 },
        mimeType: "video/mp4",
      })}
      message={create(MessageSchema, { channelId: "ch1", id: "m1" })}
    />,
    "/app/ws1/ch1",
    (routes) => {
      routes.rpc(AttachmentService.method.getDownloadUrl, getDownloadUrl);
    },
  );
  return getDownloadUrl;
};

describe("VideoAttachment", () => {
  test("サムネイルがあれば再生前に表示し、再生中の動画のポスターにも使う", async () => {
    await render({ height: 540, width: 960 });

    await waitFor(() => {
      expect(document.querySelector("img")).toHaveAttribute(
        "src",
        "https://storage.example.com/v1-thumbnail",
      );
    });

    await userEvent.click(screen.getByRole("button", { name: "demo.mp4 を再生" }));
    await waitFor(() => {
      expect(mediaPlayer.getState().track).toMatchObject({
        attachmentId: "v1",
        posterUrl: "https://storage.example.com/v1-thumbnail",
      });
    });
    expect(document.querySelector("img")).toBeNull();
  });

  test("サムネイルがなければ画像を取得しない", async () => {
    const getDownloadUrl = await render(undefined);

    expect(screen.getByText("demo.mp4")).toBeInTheDocument();
    expect(document.querySelector("img")).toBeNull();
    expect(getDownloadUrl).not.toHaveBeenCalled();
  });
});
