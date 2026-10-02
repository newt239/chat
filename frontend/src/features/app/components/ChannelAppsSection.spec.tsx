import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { AppSchema, AppService } from "#/gen/chat/v1/app_service_pb";
import { UserSummarySchema } from "#/gen/chat/v1/user_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelAppsSection } from "./ChannelAppsSection";

import type { AddAppToChannelRequest } from "#/gen/chat/v1/app_service_pb";

const deploy = create(AppSchema, {
  canManage: true,
  createdBy: create(UserSummarySchema, { displayName: "Alice", id: "u1" }),
  id: "a1",
  name: "デプロイ通知",
});
const official = create(AppSchema, { id: "a0", isOfficial: true, name: "Chat" });
const monitor = create(AppSchema, {
  canManage: true,
  createdBy: create(UserSummarySchema, { displayName: "Bob", id: "u2" }),
  id: "a2",
  name: "監視",
});

const setup = async () => {
  const add = vi.fn<(req: AddAppToChannelRequest) => void>();
  const rendered = await renderWithProviders(
    <ChannelAppsSection workspaceId="ws1" channelId="c1" />,
    "/app/ws1",
    (routes) => {
      routes.rpc(AppService.method.listChannelApps, () => ({ apps: [deploy] }));
      routes.rpc(AppService.method.listApps, () => ({ apps: [official, deploy, monitor] }));
      routes.rpc(AppService.method.addAppToChannel, (req) => {
        add(req);
        return {};
      });
    },
  );
  await screen.findByText("デプロイ通知");
  return { ...rendered, add };
};

describe("ChannelAppsSection", () => {
  test("参加中のアプリと作成者を出し、管理できるものだけ編集と外すボタンを出す", async () => {
    await setup();
    expect(screen.getByText("Alice が作成 · 未使用")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "デプロイ通知 を編集" })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "デプロイ通知 をチャンネルから外す" }),
    ).toBeInTheDocument();
  });

  test("参加していない管理できるアプリを追加でき、新しいアプリの作成はダイアログで開く", async () => {
    const { add, router } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "アプリを追加" }));
    expect(screen.queryByRole("menuitem", { name: "Chat を追加" })).not.toBeInTheDocument();
    await userEvent.click(await screen.findByRole("menuitem", { name: "監視 を追加" }));
    await waitFor(() => {
      expect(add).toHaveBeenCalledWith(expect.objectContaining({ appId: "a2", channelId: "c1" }));
    });

    await userEvent.click(screen.getByRole("button", { name: "アプリを追加" }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "新しいアプリを作成" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ dialog: "add-app" });
    });
  });
});
