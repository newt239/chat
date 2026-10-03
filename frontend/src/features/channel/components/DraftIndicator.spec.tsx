import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { DraftSchema, DraftService } from "#/gen/chat/v1/draft_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { DraftIndicator } from "./DraftIndicator";

const render = (channelId: string) =>
  renderWithProviders(
    <DraftIndicator workspaceId="ws1" channelId={channelId} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(DraftService.method.listDrafts, () => ({
        drafts: [create(DraftSchema, { body: "書きかけ", channelId: "ch1", parentId: "m1" })],
      }));
    },
  );

describe("DraftIndicator", () => {
  test("スレッドを含め、チャンネルに下書きがあれば印を出す", async () => {
    await render("ch1");
    expect(await screen.findByRole("img", { name: "下書きあり" })).toBeInTheDocument();
  });

  test("下書きのないチャンネルには出さない", async () => {
    await render("ch2");
    expect(screen.queryByRole("img", { name: "下書きあり" })).not.toBeInTheDocument();
  });
});
