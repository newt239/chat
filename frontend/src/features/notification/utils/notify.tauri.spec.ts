import { expect, test, vi } from "vite-plus/test";

import { showNotification } from "#/features/notification/utils/notify";

const show = vi.fn();

vi.mock("#/lib/platform/platform", () => ({ isTauri: true }));
vi.mock("@tauri-apps/plugin-notification", () => ({
  isPermissionGranted: () => Promise.resolve(true),
  requestPermission: () => Promise.resolve("granted"),
  sendNotification: show,
}));

test("Tauri ではプラグインで通知を出す", async () => {
  await showNotification({ body: "本文", onClick: () => {}, tag: "m1", title: "題" });

  expect(show).toHaveBeenCalledWith({ body: "本文", title: "題" });
});
