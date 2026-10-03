import { afterEach, expect, test, vi } from "vite-plus/test";

import { setAppBadge } from "#/features/layout/utils/appBadge";

afterEach(() => {
  vi.unstubAllGlobals();
});

test("未読があれば数を、なければバッジを消す", async () => {
  const set = vi.fn(async () => {});
  const clear = vi.fn(async () => {});
  vi.stubGlobal("navigator", { clearAppBadge: clear, setAppBadge: set });

  await setAppBadge(3);
  await setAppBadge(0);

  expect(set).toHaveBeenCalledWith(3);
  expect(clear).toHaveBeenCalledOnce();
});

test("Badging API がなければ何もしない", async () => {
  vi.stubGlobal("navigator", {});

  await expect(setAppBadge(1)).resolves.toBeUndefined();
});
