import { describe, expect, test } from "vite-plus/test";

import { searchFilterValues, searchQuerySchema } from "#/features/search/schemas";

describe("searchFilterValues", () => {
  test("検索フィルタの選択肢が OpenAPI の enum と一致する", () => {
    expect([...searchFilterValues]).toEqual(["all", "messages", "channels", "users", "groups"]);
  });
});

describe("searchQuerySchema", () => {
  test("クエリが無いときは既定値を返す", () => {
    expect(searchQuerySchema.parse({})).toEqual({ filter: "all", page: 1, q: "" });
  });

  test("正しいクエリはそのまま返す", () => {
    expect(searchQuerySchema.parse({ filter: "messages", page: 3, q: "hello" })).toEqual({
      filter: "messages",
      page: 3,
      q: "hello",
    });
  });

  test("不正な filter は all にフォールバックする", () => {
    expect(searchQuerySchema.parse({ filter: "unknown" }).filter).toBe("all");
  });

  test("不正な page は 1 にフォールバックする", () => {
    expect(searchQuerySchema.parse({ page: -1 }).page).toBe(1);
  });
});
