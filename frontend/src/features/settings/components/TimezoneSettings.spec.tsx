import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { UserService } from "#/gen/chat/v1/user_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { TimezoneSettings } from "./TimezoneSettings";

import type { UserPreferences } from "#/gen/chat/v1/user_pb";

const setup = () => {
  const saved: UserPreferences[] = [];
  void renderWithProviders(<TimezoneSettings />, "/app/ws1", (routes) => {
    routes.rpc(UserService.method.updatePreferences, ({ preferences }) => {
      if (preferences) {
        saved.push(preferences);
      }
      return { preferences };
    });
  });
  return saved;
};

describe("TimezoneSettings", () => {
  test("候補からタイムゾーンを選んでアカウントに保存する", async () => {
    const saved = setup();

    await userEvent.type(await screen.findByRole("combobox"), "Asia/Tok");
    await userEvent.click(await screen.findByRole("option", { name: "Asia/Tokyo" }));

    await waitFor(() => {
      expect(saved.at(-1)?.timezone).toBe("Asia/Tokyo");
    });
  });

  test("自動更新の切り替えをアカウントに保存する", async () => {
    const saved = setup();

    await userEvent.click(
      await screen.findByRole("switch", { name: "タイムゾーンを自動で更新する" }),
    );

    await waitFor(() => {
      expect(saved.at(-1)?.timezoneAutoUpdate).toBe(true);
    });
  });
});
