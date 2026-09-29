import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import {
  ScheduledMessageSchema,
  ScheduledMessageService,
  ScheduledMessageStatus,
} from "#/gen/chat/v1/scheduled_message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ScheduledMessageList } from "./ScheduledMessageList";

const messages = [
  create(ScheduledMessageSchema, {
    body: "送信済み 1",
    id: "s1",
    scheduledAt: timestampFromDate(new Date(2026, 8, 1, 9, 0)),
    status: ScheduledMessageStatus.SENT,
  }),
  create(ScheduledMessageSchema, {
    body: "予約中",
    id: "s2",
    scheduledAt: timestampFromDate(new Date(2026, 8, 2, 9, 0)),
    status: ScheduledMessageStatus.SCHEDULED,
  }),
  create(ScheduledMessageSchema, {
    body: "送信済み 2",
    id: "s3",
    scheduledAt: timestampFromDate(new Date(2026, 8, 3, 9, 0)),
    status: ScheduledMessageStatus.SENT,
  }),
];

const render = (mode: "pending" | "sent", items = messages) =>
  renderWithProviders(
    <ScheduledMessageList workspaceId="ws1" mode={mode} />,
    "/app/ws1/drafts",
    (routes) => {
      routes.rpc(ScheduledMessageService.method.listScheduledMessages, () => ({
        scheduledMessages: items,
      }));
    },
  );

describe("ScheduledMessageList", () => {
  test("予約中のタブには送信前のものだけを出す", async () => {
    await render("pending");
    expect(await screen.findByText("予約中")).toBeInTheDocument();
    expect(screen.queryByText("送信済み 1")).not.toBeInTheDocument();
  });

  test("送信済みのタブは新しい順に並べる", async () => {
    await render("sent");
    const bodies = await screen.findAllByText(/送信済み \d/);
    expect(bodies.map((node) => node.textContent)).toEqual(["送信済み 2", "送信済み 1"]);
  });

  test("空なら案内を出す", async () => {
    await render("pending", []);
    expect(await screen.findByText("予約中のメッセージはありません")).toBeInTheDocument();
  });
});
