import { createMemoryHistory, createRouter } from "@tanstack/react-router";
import { beforeEach, describe, expect, test } from "vite-plus/test";

import { store } from "#/providers/store";
import { clearAuthAtom } from "#/providers/store/auth";
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

  test("threads は $channelId より優先してマッチする", () => {
    expect(matchLeaf("/app/ws1/threads")?.routeId).toBe("/app/$workspaceId/threads");
  });
});

describe("認証ガード", () => {
  beforeEach(() => {
    store.set(clearAuthAtom);
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
