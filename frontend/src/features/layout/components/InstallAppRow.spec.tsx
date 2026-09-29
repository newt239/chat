import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test, vi } from "vite-plus/test";

import { listenInstallPrompt } from "#/lib/installPrompt";

import { InstallAppRow } from "./InstallAppRow";

test("インストールできないときは出さない", () => {
  render(<InstallAppRow />);
  expect(screen.queryByRole("button", { name: "アプリをインストール" })).not.toBeInTheDocument();
});

test("beforeinstallprompt を受け取ったら押してインストールを促す", async () => {
  listenInstallPrompt();
  render(<InstallAppRow />);
  const prompt = vi.fn(() => Promise.resolve({ outcome: "accepted" as const }));

  act(() => {
    globalThis.dispatchEvent(Object.assign(new Event("beforeinstallprompt"), { prompt }));
  });
  await userEvent.click(screen.getByRole("button", { name: "アプリをインストール" }));

  expect(prompt).toHaveBeenCalledOnce();
  expect(screen.queryByRole("button", { name: "アプリをインストール" })).not.toBeInTheDocument();
});
