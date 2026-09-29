import { expect, test, vi } from "vite-plus/test";

import { showNotification } from "#/lib/platform/notify";

const show = vi.fn();

vi.mock("#/lib/platform/platform", () => ({ isTauri: true }));
vi.mock("#/lib/platform/tauri/notification", () => ({
  isGranted: () => Promise.resolve(true),
  request: () => Promise.resolve(true),
  show,
}));

test("Tauri ではプラグインで通知を出す", async () => {
  await showNotification({ body: "本文", onClick: () => {}, tag: "m1", title: "題" });

  expect(show).toHaveBeenCalledWith("題", "本文");
});
