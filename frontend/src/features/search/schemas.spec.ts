import { describe, expect, test } from "vite-plus/test";

import {
  searchFilterMessages,
  searchFilterValues,
  searchQuerySchema,
} from "#/features/search/schemas";
import { SearchTarget } from "#/gen/chat/v1/search_service_pb";

describe("searchFilterMessages", () => {
  test("すべての検索フィルタが proto の enum の異なる値に対応する", () => {
    const values = searchFilterValues.map((filter) => searchFilterMessages[filter]);
    expect(values).not.toContain(SearchTarget.UNSPECIFIED);
    expect(new Set(values).size).toBe(searchFilterValues.length);
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
