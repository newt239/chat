import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { UserSummarySchema } from "#/gen/chat/v1/user_pb";
import { WebhookSchema, WebhookService } from "#/gen/chat/v1/webhook_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { WebhooksSection } from "./WebhooksSection";

import type { CreateWebhookRequest } from "#/gen/chat/v1/webhook_service_pb";

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
  const createWebhook = vi.fn<(req: CreateWebhookRequest) => void>();
  await renderWithProviders(<WebhooksSection channelId="c1" />, "/app/ws1", (routes) => {
    routes.rpc(WebhookService.method.listWebhooks, () => ({ webhooks }));
    routes.rpc(WebhookService.method.createWebhook, (req) => {
      createWebhook(req);
      return {
        token: "secret-token",
        webhook: create(WebhookSchema, { canManage: true, id: "w3", name: req.name }),
      };
    });
  });
  await screen.findByText("デプロイ通知");
  return { createWebhook };
};

describe("WebhooksSection", () => {
  test("作成者と最終使用を表示し、編集できるものだけ編集ボタンを出す", async () => {
    await setup();
    expect(screen.getByText("Alice が作成 · 未使用")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "デプロイ通知 を編集" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "監視 を編集" })).not.toBeInTheDocument();
  });

  test("発行すると URL を一度だけ表示する", async () => {
    const { createWebhook } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "追加" }));
    await userEvent.type(await screen.findByRole("textbox", { name: "名前" }), " CI ");
    await userEvent.click(screen.getByRole("button", { name: "発行" }));

    await waitFor(() => {
      expect(createWebhook).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", name: "CI" }),
      );
    });
    expect(
      await screen.findByText(/\/webhooks\/w3\/secret-token$/, { selector: "code" }),
    ).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "完了" }));
    await waitFor(() => {
      expect(screen.queryByText(/secret-token/)).not.toBeInTheDocument();
    });
  });

  test("名前が空なら発行しない", async () => {
    const { createWebhook } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "追加" }));
    await userEvent.click(await screen.findByRole("button", { name: "発行" }));
    expect(await screen.findByText("名前を入力してください")).toBeInTheDocument();
    expect(createWebhook).not.toHaveBeenCalled();
  });
});
