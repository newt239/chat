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
    expect(searchQuerySchema.parse({})).toEqual({
      filter: "all",
      page: 1,
      q: "",
      replies: true,
      sort: "newest",
      subs: true,
    });
  });

  test("正しいクエリはそのまま返す", () => {
    expect(
      searchQuerySchema.parse({
        filter: "messages",
        page: 3,
        q: "hello",
        replies: false,
        sort: "relevance",
        subs: false,
      }),
    ).toEqual({
      filter: "messages",
      page: 3,
      q: "hello",
      replies: false,
      sort: "relevance",
      subs: false,
    });
  });

  test("不正な filter は all にフォールバックする", () => {
    expect(searchQuerySchema.parse({ filter: "unknown" }).filter).toBe("all");
  });

  test("不正な sort は新しい順にフォールバックする", () => {
    expect(searchQuerySchema.parse({ sort: "old" }).sort).toBe("newest");
  });

  test("不正な page は 1 にフォールバックする", () => {
    expect(searchQuerySchema.parse({ page: -1 }).page).toBe(1);
  });
});
