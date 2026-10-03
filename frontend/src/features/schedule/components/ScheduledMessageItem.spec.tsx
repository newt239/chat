import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import {
  ScheduledMessageSchema,
  ScheduledMessageService,
  ScheduledMessageStatus,
} from "#/gen/chat/v1/scheduled_message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ScheduledMessageItem } from "./ScheduledMessageItem";

import type {
  ScheduledMessage,
  SendScheduledMessageNowRequest,
  UpdateScheduledMessageRequest,
} from "#/gen/chat/v1/scheduled_message_service_pb";

const future = new Date(Date.now() + 2 * 24 * 60 * 60 * 1000);
future.setSeconds(0, 0);

const scheduled = (
  init: Partial<Pick<ScheduledMessage, "status" | "failureReason" | "sentMessageId">>,
) =>
  create(ScheduledMessageSchema, {
    attachmentIds: ["a1", "a2"],
    body: "定例の議事録です",
    channelId: "c1",
    id: "s1",
    location: { label: "会議室", latitude: 1, longitude: 2 },
    scheduledAt: timestampFromDate(future),
    status: ScheduledMessageStatus.SCHEDULED,
    ...init,
  });

type Handlers = {
  update: (req: UpdateScheduledMessageRequest) => void;
  sendNow: (req: SendScheduledMessageNowRequest) => void;
};

const render = (message: ScheduledMessage, handlers: Handlers) =>
  renderWithProviders(
    <ScheduledMessageItem workspaceId="ws1" message={message} label="#general" />,
    "/app/ws1/drafts",
    (routes) => {
      routes.rpc(ScheduledMessageService.method.updateScheduledMessage, (req) => {
        handlers.update(req);
        return {};
      });
      routes.rpc(ScheduledMessageService.method.sendScheduledMessageNow, (req) => {
        handlers.sendNow(req);
        return {};
      });
      routes.rpc(ScheduledMessageService.method.listScheduledMessages, () => ({
        scheduledMessages: [],
      }));
    },
  );

const handlers = () => ({
  sendNow: vi.fn<(req: SendScheduledMessageNowRequest) => void>(),
  update: vi.fn<(req: UpdateScheduledMessageRequest) => void>(),
});

describe("ScheduledMessageItem", () => {
  test("本文・位置情報・添付の数を出し、編集で本文と日時を更新する", async () => {
    const h = handlers();
    await render(scheduled({}), h);

    expect(screen.getByText("定例の議事録です")).toBeInTheDocument();
    expect(screen.getByText("会議室")).toBeInTheDocument();
    expect(screen.getByText("添付 2 件")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "予約の操作" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "編集" }));
    const body = screen.getByRole("textbox", { name: "本文" });
    await userEvent.clear(body);
    await userEvent.type(body, "修正版");
    await userEvent.click(screen.getByRole("button", { name: "予約する" }));

    await waitFor(() => {
      expect(h.update).toHaveBeenCalledWith(expect.objectContaining({ body: "修正版", id: "s1" }));
    });
  });

  test("失敗した予約は理由を出し、今すぐ送信できる", async () => {
    const h = handlers();
    await render(
      scheduled({
        failureReason: "操作を実行する権限がありません",
        status: ScheduledMessageStatus.FAILED,
      }),
      h,
    );

    expect(screen.getByText("送信失敗")).toBeInTheDocument();
    expect(screen.getByText("操作を実行する権限がありません")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "予約の操作" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "今すぐ送信" }));
    await waitFor(() => {
      expect(h.sendNow).toHaveBeenCalledWith(expect.objectContaining({ id: "s1" }));
    });
  });

  test("送信済みは投稿へのリンクを出し、編集できない", async () => {
    await render(
      scheduled({ sentMessageId: "m9", status: ScheduledMessageStatus.SENT }),
      handlers(),
    );

    expect(screen.getByRole("link", { name: "メッセージを表示" })).toHaveAttribute(
      "href",
      "/app/ws1/c1?message=m9",
    );
    await userEvent.click(screen.getByRole("button", { name: "予約の操作" }));
    expect(screen.queryByRole("menuitem", { name: "編集" })).not.toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "削除" })).toBeInTheDocument();
  });
});
