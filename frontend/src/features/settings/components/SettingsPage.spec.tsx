import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { SettingsPage } from "./SettingsPage";

describe("SettingsPage", () => {
  test("URL の項目を開き、左の項目で切り替える", async () => {
    const { router } = await renderWithProviders(
      <SettingsPage />,
      "/app/ws1/settings/shortcuts",
      () => {},
    );
    expect(await screen.findByRole("heading", { name: "ショートカット" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ショートカット" })).toHaveAttribute(
      "aria-current",
      "page",
    );

    await userEvent.click(screen.getByRole("link", { name: "テーマ" }));
    expect(router.state.location.pathname).toBe("/app/ws1/settings/theme");
    expect(await screen.findByRole("heading", { name: "テーマ" })).toBeInTheDocument();
  });
});
