import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { ensureSession, refreshOrSignOut } from "#/lib/session";
import { sessionAtom } from "#/providers/store/auth";
import { store } from "#/providers/store/store";

vi.mock("#/lib/navigation", () => ({ navigateTo: vi.fn() }));

const respondRefresh = (response: Response) =>
  vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(response.clone()));

const unauthenticated = Response.json(
  { code: "unauthenticated", message: "no session" },
  { status: 401 },
);

describe("session", () => {
  beforeEach(() => {
    store.set(sessionAtom, null);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("同時に呼ばれた refresh は 1 回の通信にまとめ、セッションを始める", async () => {
    const fetchMock = respondRefresh(Response.json({ accessToken: "new", user: { id: "u1" } }));
    const tokens = await Promise.all([refreshOrSignOut(), refreshOrSignOut()]);
    expect(tokens).toEqual(["new", "new"]);
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(store.get(sessionAtom)).toEqual({ accessToken: "new", userId: "u1" });
  });

  test("Cookie で取り直せなければ ensureSession は false を返す", async () => {
    respondRefresh(unauthenticated);
    await expect(ensureSession()).resolves.toBe(false);
  });

  test("認証切れならセッションを消し、ほかの失敗ではセッションを残す", async () => {
    store.set(sessionAtom, { accessToken: "old", userId: "u1" });
    respondRefresh(Response.json({ code: "unavailable", message: "down" }, { status: 503 }));
    await expect(refreshOrSignOut()).rejects.toThrow();
    expect(store.get(sessionAtom)).not.toBeNull();

    vi.restoreAllMocks();
    respondRefresh(unauthenticated);
    await expect(refreshOrSignOut()).rejects.toThrow();
    expect(store.get(sessionAtom)).toBeNull();
  });
});
