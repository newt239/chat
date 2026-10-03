import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";
import { MessageAttachmentSchema } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AudioAttachment } from "./AudioAttachment";

describe("AudioAttachment", () => {
  test("署名付き URL をブラウザのプレイヤーで再生する", async () => {
    await renderWithProviders(
      <AudioAttachment
        attachment={create(MessageAttachmentSchema, {
          fileName: "voice.m4a",
          id: "a1",
          mimeType: "audio/mp4",
        })}
      />,
      "/app/ws1/ch1",
      (routes) => {
        routes.rpc(AttachmentService.method.getDownloadUrl, () => ({
          url: "https://storage.example.com/voice",
        }));
      },
    );

    expect(screen.getByText("voice.m4a")).toBeInTheDocument();
    await waitFor(() => {
      expect(document.querySelector("audio")).toHaveAttribute(
        "src",
        "https://storage.example.com/voice",
      );
    });
  });
});
