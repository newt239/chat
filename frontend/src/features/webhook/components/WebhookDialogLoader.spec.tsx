import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { WebhookSchema, WebhookService } from "#/gen/chat/v1/webhook_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { WebhookDialogLoader } from "./WebhookDialogLoader";

import type { CreateWebhookRequest } from "#/gen/chat/v1/webhook_service_pb";

const webhooks = [
  create(WebhookSchema, { canManage: true, id: "w1", name: "デプロイ通知" }),
  create(WebhookSchema, { canManage: false, id: "w2", name: "監視" }),
];

const render = async (webhookId: string | null) => {
  const createWebhook = vi.fn<(req: CreateWebhookRequest) => void>();
  const onClose = vi.fn<() => void>();
  await renderWithProviders(
    <WebhookDialogLoader channelId="c1" webhookId={webhookId} onClose={onClose} />,
    "/app/ws1/c1",
    (routes) => {
      routes.rpc(WebhookService.method.listWebhooks, () => ({ webhooks }));
      routes.rpc(WebhookService.method.createWebhook, (req) => {
        createWebhook(req);
        return {
          token: "secret-token",
          webhook: create(WebhookSchema, { canManage: true, id: "w3", name: req.name }),
        };
      });
    },
  );
  return { createWebhook, onClose };
};

describe("WebhookDialogLoader", () => {
  test("ID の Webhook を読み込んでから編集のダイアログを開く", async () => {
    await render("w1");
    expect(await screen.findByRole("textbox", { name: "名前" })).toHaveValue("デプロイ通知");
  });

  test("見つからない Webhook や編集できない Webhook には開かない", async () => {
    await render("missing");
    await render("w2");
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).toBeNull();
    });
  });

  test("発行すると URL を一度だけ表示し、完了で閉じる", async () => {
    const { createWebhook, onClose } = await render(null);
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
    expect(onClose).toHaveBeenCalled();
  });

  test("名前が空なら発行しない", async () => {
    const { createWebhook } = await render(null);
    await userEvent.click(await screen.findByRole("button", { name: "発行" }));
    expect(await screen.findByText("名前を入力してください")).toBeInTheDocument();
    expect(createWebhook).not.toHaveBeenCalled();
  });
});
