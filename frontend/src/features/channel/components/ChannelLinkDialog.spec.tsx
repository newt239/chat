import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelLinkSchema } from "#/gen/chat/v1/channel_link_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelLinkDialog } from "./ChannelLinkDialog";

describe("ChannelLinkDialog", () => {
  test("編集するリンクの値を入力欄の初期値にする", async () => {
    await renderWithProviders(
      <ChannelLinkDialog
        channelId="c1"
        link={create(ChannelLinkSchema, { id: "l1", title: "設計書", url: "https://a.com" })}
        onClose={vi.fn<() => void>()}
      />,
      "/app/ws1/c1",
      () => {},
    );
    expect(await screen.findByRole("textbox", { name: "表示名" })).toHaveValue("設計書");
    expect(screen.getByRole("textbox", { name: "URL" })).toHaveValue("https://a.com");
  });
});
