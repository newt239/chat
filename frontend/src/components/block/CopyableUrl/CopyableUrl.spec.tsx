import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { CopyableUrl } from "./CopyableUrl";

describe("CopyableUrl", () => {
  test("URL を表示し、コピーするとクリップボードに入る", async () => {
    const user = userEvent.setup();
    await renderWithProviders(
      <CopyableUrl url="https://example.com/join/ws1" />,
      "/app/ws1",
      () => {},
    );

    expect(screen.getByText("https://example.com/join/ws1")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "コピー" }));
    await waitFor(async () => {
      expect(await navigator.clipboard.readText()).toBe("https://example.com/join/ws1");
    });
  });
});
