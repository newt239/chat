import { screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { ErrorPage } from "./ErrorPage";

describe("ErrorPage", () => {
  test("ページが見つからないときはその旨を表示する", async () => {
    await renderWithProviders(<ErrorPage isNotFound routeId="__root__" />, "/app", () => {});

    expect(screen.getByRole("heading", { name: "404" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "トップへ戻る" })).toHaveAttribute("href", "/app");
  });

  test("エラーのときはエラーの旨を表示する", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    await renderWithProviders(
      <ErrorPage error={new Error("boom")} reset={() => {}} />,
      "/app",
      () => {},
    );

    expect(screen.getByRole("heading", { name: "問題が発生しました" })).toBeInTheDocument();
  });
});
