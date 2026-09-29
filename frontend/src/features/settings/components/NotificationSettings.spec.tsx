import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { NotificationService, PushPlatform } from "#/gen/chat/v1/notification_service_pb";
import { NotificationLevel } from "#/gen/chat/v1/user_pb";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { notificationPreferencesAtom } from "#/providers/store/notificationPreferences";
import { renderWithProviders } from "#/test/renderWithProviders";

import { NotificationSettings } from "./NotificationSettings";

import type { RegisterPushTokenRequest } from "#/gen/chat/v1/notification_service_pb";
import type { UserPreferences } from "#/gen/chat/v1/user_pb";

const push = vi.hoisted(() => ({ supported: false }));

vi.mock("#/features/settings/utils/pushMessaging", () => ({
  isPushSupported: () => push.supported,
  registerPush: vi.fn(() => Promise.resolve("token-1")),
  unregisterPush: vi.fn(() => Promise.resolve()),
}));

const setup = () => {
  const saved: UserPreferences[] = [];
  const registered: RegisterPushTokenRequest[] = [];
  const rendered = renderWithProviders(<NotificationSettings />, "/app/ws1", (routes) => {
    routes.rpc(UserService.method.updatePreferences, ({ preferences }) => {
      if (preferences) {
        saved.push(preferences);
      }
      return { preferences };
    });
    routes.rpc(NotificationService.method.registerPushToken, (req) => {
      registered.push(req);
      return {};
    });
    routes.rpc(NotificationService.method.unregisterPushToken, () => ({}));
  });
  return { registered, rendered, saved };
};

describe("NotificationSettings", () => {
  beforeEach(() => {
    vi.stubGlobal("Notification", { requestPermission: vi.fn(() => Promise.resolve("granted")) });
  });

  afterEach(() => {
    push.supported = false;
    localStorage.clear();
    vi.stubGlobal("Notification", undefined);
  });

  test("通知する範囲はアカウントに保存する", async () => {
    const { saved } = setup();

    await userEvent.click(await screen.findByRole("radio", { name: "すべて" }));

    await waitFor(() => {
      expect(saved.at(-1)?.notificationLevel).toBe(NotificationLevel.ALL);
    });
  });

  test("Firebase が使えないときはプッシュ通知の項目を出さない", async () => {
    setup();
    await screen.findByRole("radio", { name: "すべて" });
    expect(screen.queryByRole("switch", { name: "プッシュ通知" })).not.toBeInTheDocument();
  });

  test("プッシュ通知をオンにするとトークンを登録して端末に覚える", async () => {
    push.supported = true;
    const { registered, rendered } = setup();

    await userEvent.click(await screen.findByRole("switch", { name: "プッシュ通知" }));

    await waitFor(() => {
      expect(registered).toMatchObject([{ platform: PushPlatform.WEB, token: "token-1" }]);
    });
    const { store } = await rendered;
    expect(store.get(notificationPreferencesAtom).pushToken).toBe("token-1");
  });
});
