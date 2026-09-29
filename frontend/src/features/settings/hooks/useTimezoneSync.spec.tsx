import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ToastRegion } from "#/components/ui/ToastRegion";
import { UserPreferencesSchema } from "#/gen/chat/v1/user_pb";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { useSyncPreferences } from "./usePreferences";
import { useTimezoneSync } from "./useTimezoneSync";

import type { UserPreferences } from "#/gen/chat/v1/user_pb";

const deviceTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
const otherTimezone = deviceTimezone === "Asia/Tokyo" ? "Europe/London" : "Asia/Tokyo";

// アプリと同じく、アカウントの設定を読み込んだ後に確かめる
const Probe = () => {
  useSyncPreferences();
  useTimezoneSync();
  return <ToastRegion />;
};

const setup = (saved: Pick<UserPreferences, "timezone" | "timezoneAutoUpdate">) => {
  const updates: UserPreferences[] = [];
  void renderWithProviders(<Probe />, "/app/ws1", (routes) => {
    routes.rpc(UserService.method.getMe, () => ({
      user: { id: "u1", preferences: create(UserPreferencesSchema, { locale: "ja", ...saved }) },
    }));
    routes.rpc(UserService.method.updatePreferences, ({ preferences }) => {
      if (preferences) {
        updates.push(preferences);
      }
      return { preferences };
    });
  });
  return updates;
};

describe("useTimezoneSync", () => {
  test("未設定なら尋ねずに端末のタイムゾーンを保存する", async () => {
    const updates = setup({ timezone: "", timezoneAutoUpdate: false });

    await waitFor(() => {
      expect(updates.at(-1)?.timezone).toBe(deviceTimezone);
    });
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  test("自動更新が有効なら尋ねずに更新する", async () => {
    const updates = setup({ timezone: otherTimezone, timezoneAutoUpdate: true });

    await waitFor(() => {
      expect(updates.at(-1)?.timezone).toBe(deviceTimezone);
    });
    expect(updates.at(-1)?.timezoneAutoUpdate).toBe(true);
  });

  test("自動更新が無効ならトーストで尋ね、押したときだけ更新する", async () => {
    const updates = setup({ timezone: otherTimezone, timezoneAutoUpdate: false });

    await userEvent.click(await screen.findByRole("button", { name: "更新する" }));

    await waitFor(() => {
      expect(updates.at(-1)?.timezone).toBe(deviceTimezone);
    });
  });
});
