import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { DraftSchema, DraftService } from "#/gen/chat/v1/draft_service_pb";
import { ScheduledMessageService } from "#/gen/chat/v1/scheduled_message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { DraftsPage } from "./DraftsPage";

describe("DraftsPage", () => {
  test("下書き・予約済み・送信済みをタブで切り替える", async () => {
    await renderWithProviders(<DraftsPage />, "/app/ws1/drafts", (routes) => {
      routes.rpc(DraftService.method.listDrafts, () => ({
        drafts: [create(DraftSchema, { body: "書きかけ", channelId: "c1", id: "d1" })],
      }));
      routes.rpc(ScheduledMessageService.method.listScheduledMessages, () => ({
        scheduledMessages: [],
      }));
    });

    expect(screen.getByRole("heading", { name: "下書きと予約" })).toBeInTheDocument();
    expect(await screen.findByText("書きかけ")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("tab", { name: "予約済み" }));
    expect(await screen.findByText("予約中のメッセージはありません")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("tab", { name: "送信済み" }));
    expect(
      await screen.findByText("予約から送信したメッセージはまだありません"),
    ).toBeInTheDocument();
  });
});
