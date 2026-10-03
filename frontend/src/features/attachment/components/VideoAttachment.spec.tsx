import { create } from "@bufbuild/protobuf";
import { waitFor } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";
import { MessageAttachmentSchema } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { VideoAttachment } from "./VideoAttachment";

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
    />,
    "/app/ws1/ch1",
    (routes) => {
      routes.rpc(AttachmentService.method.getDownloadUrl, getDownloadUrl);
    },
  );
  return getDownloadUrl;
};

describe("VideoAttachment", () => {
  test("サムネイルがあれば再生前のポスターにする", async () => {
    await render({ height: 540, width: 960 });

    await waitFor(() => {
      expect(document.querySelector("video")).toHaveAttribute(
        "poster",
        "https://storage.example.com/v1-thumbnail",
      );
    });
    expect(document.querySelector("video")).toHaveAttribute(
      "src",
      "https://storage.example.com/v1",
    );
  });

  test("サムネイルがなければポスターを取得しない", async () => {
    const getDownloadUrl = await render(undefined);

    await waitFor(() => {
      expect(document.querySelector("video")).toHaveAttribute(
        "src",
        "https://storage.example.com/v1",
      );
    });
    expect(document.querySelector("video")).not.toHaveAttribute("poster");
    expect(getDownloadUrl).toHaveBeenCalledOnce();
  });
});
