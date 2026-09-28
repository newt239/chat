import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { RightSidePanel } from "./RightSidePanel";

describe("RightSidePanel", () => {
  test("?panel= で表示中のチャンネルのパネルを開き、閉じると search から消す", async () => {
    const { router } = await renderWithProviders(
      <RightSidePanel workspaceId="ws1" />,
      "/app/ws1/c1?panel=pins&message=m1",
      () => {},
    );
    expect(screen.getByRole("complementary", { name: "ピン留め" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "閉じる" }));
    expect(router.state.location.search).toEqual({ message: "m1" });
    await waitFor(() => {
      expect(screen.queryByRole("complementary")).toBeNull();
    });
  });

  test("チャンネルの外では ?panel= を無視し、?profile= は開く", async () => {
    await renderWithProviders(
      <RightSidePanel workspaceId="ws1" />,
      "/app/ws1?panel=pins",
      () => {},
    );
    expect(screen.queryByRole("complementary")).toBeNull();

    await renderWithProviders(
      <RightSidePanel workspaceId="ws1" />,
      "/app/ws1?profile=u-bob",
      () => {},
    );
    expect(screen.getByRole("complementary", { name: "プロフィール" })).toBeInTheDocument();
  });
});
