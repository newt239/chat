import { create } from "@bufbuild/protobuf";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { AppPermission, AppSchema, AppService } from "#/gen/chat/v1/app_service_pb";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AppDialog } from "./AppDialog";

import type {
  CreateAppRequest,
  DeleteAppRequest,
  UpdateAppRequest,
} from "#/gen/chat/v1/app_service_pb";

const app = create(AppSchema, {
  avatarUrl: "https://example.com/icon.png",
  canManage: true,
  defaultChannelId: "c1",
  id: "a1",
  name: "デプロイ通知",
  outgoingSecret: "secret-1",
  outgoingUrl: "https://example.com/hook",
  permissions: [AppPermission.POST_JOINED_CHANNELS, AppPermission.OUTGOING_WEBHOOK],
});

const setup = async (target: typeof app | null) => {
  const createApp = vi.fn<(req: CreateAppRequest) => void>();
  const update = vi.fn<(req: UpdateAppRequest) => void>();
  const remove = vi.fn<(req: DeleteAppRequest) => void>();
  const onClose = vi.fn<() => void>();
  await renderWithProviders(
    <AppDialog workspaceId="ws1" app={target} initialChannelId="c1" onClose={onClose} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [{ id: "c1", name: "general" }],
      }));
      routes.rpc(AppService.method.createApp, (req) => {
        createApp(req);
        return { app: { id: "a2" }, token: "first-token" };
      });
      routes.rpc(AppService.method.updateApp, (req) => {
        update(req);
        return { app };
      });
      routes.rpc(AppService.method.regenerateAppToken, () => ({ token: "new-token" }));
      routes.rpc(AppService.method.deleteApp, (req) => {
        remove(req);
        return {};
      });
    },
  );
  await screen.findByRole("dialog", { name: target ? "アプリを編集" : "アプリを作成" });
  return { createApp, onClose, remove, update };
};

describe("AppDialog", () => {
  test("権限を選んで作成し、着信 Webhook の URL を一度だけ表示する", async () => {
    const { createApp } = await setup(null);
    await userEvent.type(screen.getByRole("textbox", { name: /名前/ }), "監視");
    await userEvent.click(screen.getByRole("checkbox", { name: /スレッドに返信/ }));
    await userEvent.click(screen.getByRole("button", { name: "作成" }));

    await waitFor(() => {
      expect(createApp).toHaveBeenCalled();
    });
    const settings = createApp.mock.calls[0]?.[0].settings;
    expect(settings?.name).toBe("監視");
    expect(settings?.defaultChannelId).toBe("c1");
    expect(settings?.permissions).toEqual([
      AppPermission.POST_JOINED_CHANNELS,
      AppPermission.POST_THREAD_REPLIES,
    ]);
    expect(
      await screen.findByText(/\/webhooks\/a2\/first-token$/, { selector: "code" }),
    ).toBeInTheDocument();
  });

  test("送信 Webhook を選んだら送信先の URL を必須にする", async () => {
    const { createApp } = await setup(null);
    await userEvent.type(screen.getByRole("textbox", { name: /名前/ }), "監視");
    await userEvent.click(screen.getByRole("checkbox", { name: /投稿時に Webhook を送信/ }));
    await userEvent.click(screen.getByRole("button", { name: "作成" }));

    expect(await screen.findByText("送信先の URL を入力してください")).toBeInTheDocument();
    expect(createApp).not.toHaveBeenCalled();
  });

  test("編集では署名の秘密鍵を表示し、保存すると設定を送る", async () => {
    const { onClose, update } = await setup(app);
    expect(screen.getByText("secret-1", { selector: "code" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => {
      expect(update).toHaveBeenCalledWith(expect.objectContaining({ appId: "a1" }));
    });
    expect(update.mock.calls[0]?.[0].settings?.outgoingUrl).toBe("https://example.com/hook");
    expect(onClose).toHaveBeenCalled();
  });

  test("URL を再発行すると新しい URL を表示する", async () => {
    await setup(app);
    await userEvent.click(screen.getByRole("button", { name: "URL を再発行" }));
    expect(
      await screen.findByText(/\/webhooks\/a1\/new-token$/, { selector: "code" }),
    ).toBeInTheDocument();
  });

  test("確認してから削除する", async () => {
    const { onClose, remove } = await setup(app);
    await userEvent.click(screen.getByRole("button", { name: "削除" }));
    await userEvent.click(
      within(await screen.findByRole("alertdialog")).getByRole("button", { name: "削除する" }),
    );
    await waitFor(() => {
      expect(remove).toHaveBeenCalledWith(expect.objectContaining({ appId: "a1" }));
    });
    expect(onClose).toHaveBeenCalled();
  });
});
