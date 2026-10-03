import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { signOut } from "#/lib/session";
import { renderWithProviders } from "#/test/renderWithProviders";

import { LogoutConfirmDialog } from "./LogoutConfirmDialog";

vi.mock("#/lib/session", async (importOriginal) => ({
  ...(await importOriginal()),
  signOut: vi.fn(),
}));

describe("LogoutConfirmDialog", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  test("確定したときだけログアウトする", async () => {
    const logout = vi.fn(() => ({}));
    const onOpenChange = vi.fn<(isOpen: boolean) => void>();
    await renderWithProviders(
      <LogoutConfirmDialog isOpen onOpenChange={onOpenChange} />,
      "/app/ws1",
      (routes) => {
        routes.rpc(AuthService.method.logout, logout);
      },
    );
    const dialog = await screen.findByRole("alertdialog", { name: "ログアウトしますか？" });

    await userEvent.click(within(dialog).getByRole("button", { name: "キャンセル" }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(logout).not.toHaveBeenCalled();

    await userEvent.click(within(dialog).getByRole("button", { name: "ログアウト" }));
    await waitFor(() => {
      expect(signOut).toHaveBeenCalled();
    });
    expect(logout).toHaveBeenCalled();
  });
});
