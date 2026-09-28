import { create } from "@bufbuild/protobuf";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { WebhookSchema, WebhookService } from "#/gen/chat/v1/webhook_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { WebhookDialog } from "./WebhookDialog";

import type { DeleteWebhookRequest, UpdateWebhookRequest } from "#/gen/chat/v1/webhook_service_pb";

const webhook = create(WebhookSchema, {
  avatarUrl: "https://example.com/icon.png",
  canManage: true,
  id: "w1",
  name: "デプロイ通知",
});

const setup = async () => {
  const update = vi.fn<(req: UpdateWebhookRequest) => void>();
  const remove = vi.fn<(req: DeleteWebhookRequest) => void>();
  const onClose = vi.fn<() => void>();
  await renderWithProviders(
    <WebhookDialog channelId="c1" webhook={webhook} onClose={onClose} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(WebhookService.method.listWebhooks, () => ({ webhooks: [webhook] }));
      routes.rpc(WebhookService.method.updateWebhook, (req) => {
        update(req);
        return { webhook };
      });
      routes.rpc(WebhookService.method.regenerateWebhookToken, () => ({ token: "new-token" }));
      routes.rpc(WebhookService.method.deleteWebhook, (req) => {
        remove(req);
        return {};
      });
    },
  );
  await screen.findByRole("dialog", { name: "Webhook を編集" });
  return { onClose, remove, update };
};

describe("WebhookDialog", () => {
  test("アイコンを空にして保存するとアイコンを外す", async () => {
    const { onClose, update } = await setup();
    await userEvent.clear(screen.getByRole("textbox", { name: "アイコン画像の URL（任意）" }));
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => {
      expect(update).toHaveBeenCalledWith(
        expect.objectContaining({ name: "デプロイ通知", webhookId: "w1" }),
      );
    });
    expect(update.mock.calls[0]?.[0].avatarUrl).toBeUndefined();
    expect(onClose).toHaveBeenCalled();
  });

  test("アイコンが http(s) の URL でなければ保存しない", async () => {
    const { update } = await setup();
    const avatar = screen.getByRole("textbox", { name: "アイコン画像の URL（任意）" });
    await userEvent.clear(avatar);
    await userEvent.type(avatar, "ftp://example.com/a.png");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    expect(
      await screen.findByText("https:// から始まる URL を入力してください"),
    ).toBeInTheDocument();
    expect(update).not.toHaveBeenCalled();
  });

  test("URL を再発行すると新しい URL を表示する", async () => {
    await setup();
    await userEvent.click(screen.getByRole("button", { name: "URL を再発行" }));
    expect(
      await screen.findByText(/\/webhooks\/w1\/new-token$/, { selector: "code" }),
    ).toBeInTheDocument();
  });

  test("確認してから削除する", async () => {
    const { onClose, remove } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "削除" }));
    await userEvent.click(
      within(await screen.findByRole("alertdialog")).getByRole("button", { name: "削除する" }),
    );
    await waitFor(() => {
      expect(remove).toHaveBeenCalledWith(expect.objectContaining({ webhookId: "w1" }));
    });
    expect(onClose).toHaveBeenCalled();
  });
});
