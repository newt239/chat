import { createMemoryHistory, createRouter } from "@tanstack/react-router";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { sessionAtom } from "#/providers/store/auth";
import { store } from "#/providers/store/store";
import { routeTree } from "#/routeTree.gen";

const createTestRouter = (path: string) =>
  createRouter({ history: createMemoryHistory({ initialEntries: [path] }), routeTree });

const matchLeaf = (pathname: string) => createTestRouter(pathname).matchRoutes(pathname).at(-1);

describe("routeTree", () => {
  test("ワークスペース一覧にマッチする", () => {
    expect(matchLeaf("/app")?.routeId).toBe("/app/");
  });

  test("ワークスペース直下はチャンネル未選択の案内にマッチする", () => {
    const leaf = matchLeaf("/app/ws1");
    expect(leaf?.routeId).toBe("/app/$workspaceId/");
    expect(leaf?.params).toEqual({ workspaceId: "ws1" });
  });

  test("チャンネルにマッチし両方のパラメータを取り出せる", () => {
    const leaf = matchLeaf("/app/ws1/ch1");
    expect(leaf?.routeId).toBe("/app/$workspaceId/$channelId");
    expect(leaf?.params).toEqual({ channelId: "ch1", workspaceId: "ws1" });
  });

  test("search は $channelId より優先してマッチする", () => {
    expect(matchLeaf("/app/ws1/search")?.routeId).toBe("/app/$workspaceId/search");
  });

  test.each(["threads", "mentions", "bookmarks", "dms", "activity", "me", "insights", "admin"])(
    "%s は $channelId より優先してマッチする",
    (name) => {
      expect(matchLeaf(`/app/ws1/${name}`)?.routeId).toBe(`/app/$workspaceId/${name}`);
    },
  );

  test("スレッドにマッチし、チャンネルとスレッドのパラメータを取り出せる", () => {
    const leaf = matchLeaf("/app/ws1/ch1/thread/m1");
    expect(leaf?.routeId).toBe("/app/$workspaceId/$channelId/thread/$messageId");
    expect(leaf?.params).toEqual({ channelId: "ch1", messageId: "m1", workspaceId: "ws1" });
  });
});

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
