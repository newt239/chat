import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { NavigationShortcuts } from "./NavigationShortcuts";

describe("NavigationShortcuts", () => {
  test("固定の移動先を名前の見える行で並べる", async () => {
    await renderWithProviders(<NavigationShortcuts workspaceId="ws1" />, "/app/ws1", () => {});

    expect(screen.getByRole("link", { name: "スレッド" })).toHaveAttribute(
      "href",
      "/app/ws1/threads",
    );
    expect(screen.getByRole("link", { name: "下書きと予約" })).toHaveAttribute(
      "href",
      "/app/ws1/drafts",
    );
  });
});
