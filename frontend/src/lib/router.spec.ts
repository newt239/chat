import { createMemoryHistory, createRouter } from "@tanstack/react-router";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { sessionAtom } from "#/providers/store/auth";
import { store } from "#/providers/store/store";
import { routeTree } from "#/routeTree.gen";

const createTestRouter = (path: string) =>
  createRouter({ history: createMemoryHistory({ initialEntries: [path] }), routeTree });

describe("認証ガード", () => {
  beforeEach(() => {
    store.set(sessionAtom, null);
    // Cookie のリフレッシュトークンがない状態として、Refresh に Unauthenticated を返す
    vi.spyOn(globalThis, "fetch").mockImplementation(() =>
      Promise.resolve(
        Response.json({ code: "unauthenticated", message: "no session" }, { status: 401 }),
      ),
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  test.each(["/", "/app", "/app/ws1/ch1"])(
    "未認証で %s を開くとログイン画面へ遷移する",
    async (path) => {
      const router = createTestRouter(path);
      await router.load();
      expect(router.state.location.pathname).toBe("/login");
    },
  );
});
