import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserService } from "#/gen/chat/v1/user_service_pb";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { TopBar } from "./TopBar";

describe("TopBar", () => {
  test("移動のボタンでクイック切替を開く", async () => {
    const { router } = await renderWithProviders(
      <TopBar workspaceId="ws1" />,
      "/app/ws1",
      () => {},
    );

    await userEvent.click(screen.getByRole("link", { name: /チャンネルや DM に移動/ }));

    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ dialog: "switcher" });
    });
  });

  test("自分のメニューからプロフィールと設定へ移れる", async () => {
    await renderWithProviders(<TopBar workspaceId="ws1" />, "/app/ws1", (routes) => {
      routes.rpc(UserService.method.getMe, () => ({ user: currentUser }));
    });

    await userEvent.click(await screen.findByRole("button", { name: "自分のメニュー" }));

    expect(screen.getByRole("menuitem", { name: "プロフィール" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "設定" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "ログアウト" })).toBeInTheDocument();
  });
});
