import { matchRoutes } from "react-router";
import { describe, expect, test } from "vite-plus/test";

import { SearchPage } from "#/features/search/components/SearchPage";
import { ThreadListPage } from "#/features/thread/components/ThreadListPage";
import { WorkspaceSelection } from "#/features/workspace/components/WorkspaceSelection";
import { ChannelPage } from "#/routes/ChannelPage";
import { routeTree } from "#/routes/routeTree";
import { WorkspaceIndexPage } from "#/routes/WorkspaceIndexPage";

const matchLeaf = (pathname: string) => matchRoutes(routeTree, pathname)?.at(-1);

describe("routeTree", () => {
  test("ワークスペース一覧にマッチする", () => {
    expect(matchLeaf("/app")?.route.Component).toBe(WorkspaceSelection);
  });

  test("ワークスペース直下はチャンネル未選択の案内にマッチする", () => {
    const leaf = matchLeaf("/app/ws1");
    expect(leaf?.route.Component).toBe(WorkspaceIndexPage);
    expect(leaf?.params.workspaceId).toBe("ws1");
  });

  test("チャンネルにマッチし両方のパラメータを取り出せる", () => {
    const leaf = matchLeaf("/app/ws1/ch1");
    expect(leaf?.route.Component).toBe(ChannelPage);
    expect(leaf?.params.workspaceId).toBe("ws1");
    expect(leaf?.params.channelId).toBe("ch1");
  });

  // 静的セグメントが動的セグメント（:channelId）より優先されることを保証する。
  // 優先順位が崩れると /app/:ws/search が「search という名前のチャンネル」として扱われる。
  test("search は :channelId より優先してマッチする", () => {
    const leaf = matchLeaf("/app/ws1/search");
    expect(leaf?.route.Component).toBe(SearchPage);
    expect(leaf?.params.channelId).toBeUndefined();
  });

  test("threads は :channelId より優先してマッチする", () => {
    const leaf = matchLeaf("/app/ws1/threads");
    expect(leaf?.route.Component).toBe(ThreadListPage);
    expect(leaf?.params.channelId).toBeUndefined();
  });

  test("/app 配下には認証ガードの loader が 1 つだけ定義されている", () => {
    const matches = matchRoutes(routeTree, "/app/ws1/ch1") ?? [];
    const guarded = matches.filter((match) => match.route.loader !== undefined);
    expect(guarded).toHaveLength(1);
    expect(guarded[0]?.route.path).toBe("app");
  });
});
