import { Code, ConnectError } from "@connectrpc/connect";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { toast } from "#/components/ui/ToastRegion/toast";
import { notificationPreferencesAtom } from "#/features/notification/atoms";
import { NotificationService } from "#/gen/chat/v1/notification_service_pb";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { signOut } from "#/lib/session";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { AccountSettings } from "./AccountSettings";

vi.mock("#/components/ui/ToastRegion/toast", () => ({ toast: vi.fn() }));
vi.mock("#/lib/session", async (importOriginal) => ({
  ...(await importOriginal()),
  signOut: vi.fn(),
}));
vi.mock("#/features/notification/utils/pushMessaging", () => ({
  isPushSupported: () => true,
  registerPush: vi.fn(() => Promise.resolve("token-1")),
  unregisterPush: vi.fn(() => Promise.resolve()),
}));

const setup = async (deleteMe: () => object) => {
  const unregistered: string[] = [];
  const { store } = await renderWithProviders(<AccountSettings />, "/app/ws1", (routes) => {
    routes.rpc(UserService.method.getMe, () => ({ user: currentUser }));
    routes.rpc(UserService.method.deleteMe, deleteMe);
    routes.rpc(NotificationService.method.unregisterPushToken, ({ token }) => {
      unregistered.push(token);
      return {};
    });
  });
  store.set(notificationPreferencesAtom, { desktop: false, pushToken: "token-1" });
  await userEvent.click(await screen.findByRole("button", { name: "削除" }));
  await userEvent.click(
    within(await screen.findByRole("alertdialog")).getByRole("button", { name: "削除" }),
  );
  return { unregistered };
};

describe("AccountSettings", () => {
  afterEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  test("削除を拒否されたら理由を出し、プッシュ通知は解除しない", async () => {
    const { unregistered } = await setup(() => {
      throw new ConnectError("ワークスペースのオーナーは退会できません", Code.FailedPrecondition);
    });

    await waitFor(() => {
      expect(toast).toHaveBeenCalledWith(
        expect.stringContaining("ワークスペースのオーナーは退会できません"),
        { tone: "danger" },
      );
    });
    await waitFor(() => {
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    });
    expect(unregistered).toEqual([]);
    expect(signOut).not.toHaveBeenCalled();
  });

  test("削除できたらプッシュ通知を解除してからログアウトする", async () => {
    const { unregistered } = await setup(() => ({}));

    await waitFor(() => {
      expect(signOut).toHaveBeenCalled();
    });
    expect(unregistered).toEqual(["token-1"]);
  });
});
