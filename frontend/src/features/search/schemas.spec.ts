import { describe, expect, test } from "vite-plus/test";

import { searchFilterValues } from "#/features/search/schemas";

describe("searchFilterValues", () => {
  test("検索フィルタの選択肢が OpenAPI の enum と一致する", () => {
    expect([...searchFilterValues]).toEqual(["all", "messages", "channels", "users", "groups"]);
  });
});
