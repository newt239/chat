import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { SidebarNavTabs } from "./SidebarNavTabs";

describe("SidebarNavTabs", () => {
  test("固定の移動先を名前付きのリンクで並べ、表示中のページを示す", async () => {
    await renderWithProviders(<SidebarNavTabs workspaceId="ws1" />, "/app/ws1/drafts", () => {});

    const nav = screen.getByRole("navigation", { name: "移動先" });
    expect(nav).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "スレッド" })).toHaveAttribute(
      "href",
      "/app/ws1/threads",
    );
    expect(screen.getByRole("link", { name: "メンション" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ブックマーク" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "下書きと予約" })).toHaveAttribute(
      "data-status",
      "active",
    );
  });
});
