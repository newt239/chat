import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserSummarySchema } from "#/gen/chat/v1/user_pb";
import { WebhookSchema, WebhookService } from "#/gen/chat/v1/webhook_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { WebhooksSection } from "./WebhooksSection";

const webhooks = [
  create(WebhookSchema, {
    canManage: true,
    createdBy: create(UserSummarySchema, { displayName: "Alice", id: "u1" }),
    id: "w1",
    name: "デプロイ通知",
  }),
  create(WebhookSchema, {
    canManage: false,
    createdBy: create(UserSummarySchema, { displayName: "Bob", id: "u2" }),
    id: "w2",
    name: "監視",
  }),
];

const setup = async () => {
  const rendered = await renderWithProviders(
    <WebhooksSection channelId="c1" />,
    "/app/ws1",
    (routes) => {
      routes.rpc(WebhookService.method.listWebhooks, () => ({ webhooks }));
    },
  );
  await screen.findByText("デプロイ通知");
  return rendered;
};

describe("WebhooksSection", () => {
  test("作成者と最終使用を表示し、編集できるものだけ編集ボタンを出す", async () => {
    await setup();
    expect(screen.getByText("Alice が作成 · 未使用")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "デプロイ通知 を編集" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "監視 を編集" })).not.toBeInTheDocument();
  });

  test("追加と編集は URL のダイアログとして開く", async () => {
    const { router } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "追加" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ dialog: "add-webhook" });
    });
    await userEvent.click(screen.getByRole("button", { name: "デプロイ通知 を編集" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ dialog: "edit-webhook", webhook: "w1" });
    });
  });
});
