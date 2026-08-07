import { describe, expect, test } from "vite-plus/test";

import { paths } from "#/lib/paths";

describe("paths", () => {
  test("静的なパスを組み立てる", () => {
    expect(paths.login()).toBe("/login");
    expect(paths.register()).toBe("/register");
    expect(paths.app()).toBe("/app");
  });

  test("ワークスペースとスレッドのパスを組み立てる", () => {
    expect(paths.workspace("ws1")).toBe("/app/ws1");
    expect(paths.threads("ws1")).toBe("/app/ws1/threads");
  });

  test("チャンネルのパスは messageId を省略するとクエリを付けない", () => {
    expect(paths.channel("ws1", "ch1")).toBe("/app/ws1/ch1");
  });

  test("チャンネルのパスは messageId を渡すと message クエリを付ける", () => {
    expect(paths.channel("ws1", "ch1", "msg1")).toBe("/app/ws1/ch1?message=msg1");
  });

  test("検索のパスは指定した条件だけをクエリに含める", () => {
    expect(paths.search("ws1")).toBe("/app/ws1/search");
    expect(paths.search("ws1", { q: "hello" })).toBe("/app/ws1/search?q=hello");
    expect(paths.search("ws1", { q: "hello", filter: "messages", page: 2 })).toBe(
      "/app/ws1/search?q=hello&filter=messages&page=2",
    );
  });

  test("検索のパスは空文字のキーワードをクエリに含めない", () => {
    expect(paths.search("ws1", { q: "", filter: "all" })).toBe("/app/ws1/search?filter=all");
  });
});
