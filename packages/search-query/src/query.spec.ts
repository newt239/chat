import { describe, expect, it } from "vite-plus/test";

import {
  emptySearchQuery,
  formatSearchQuery,
  hasSearchConditions,
  parseSearchQuery,
  searchDateRange,
} from "./query";

describe("parseSearchQuery", () => {
  it("修飾子と語に分ける", () => {
    expect(
      parseSearchQuery("リリース from:@alice in:#dev/frontend has:image after:2026-09-01 手順"),
    ).toEqual({
      ...emptySearchQuery,
      after: "2026-09-01",
      from: ["alice"],
      has: ["image"],
      in: ["dev/frontend"],
      keywords: ["リリース", "手順"],
    });
  });

  it("@ と # は省略でき、大文字の修飾子も受け付ける", () => {
    expect(parseSearchQuery("FROM:bob IN:general HAS:Video IS:Pinned")).toMatchObject({
      from: ["bob"],
      has: ["video"],
      in: ["general"],
      is: ["pinned"],
      keywords: [],
    });
  });

  it("has:location で位置情報を共有したメッセージに絞り込む", () => {
    expect(parseSearchQuery("集合 has:location")).toMatchObject({
      has: ["location"],
      keywords: ["集合"],
    });
  });

  it("同じ条件は重ねず、日付は後のものを使う", () => {
    expect(
      parseSearchQuery("has:file has:file is:thread is:mention after:2026-09-01 after:2026-09-05"),
    ).toMatchObject({ after: "2026-09-05", has: ["file"], is: ["thread", "mention"] });
  });

  it("クォートで空白を含む名前を指定できる", () => {
    expect(parseSearchQuery('from:@"Alice Smith" "foo bar"')).toMatchObject({
      from: ["Alice Smith"],
      keywords: ['"foo bar"'],
    });
  });

  it("日付は YYYY-MM-DD だけを受け付け、それ以外は不正な日付として分ける", () => {
    expect(
      parseSearchQuery("after:2026-09-01 before:2026-02-30 after:2026/09/01 before:today"),
    ).toMatchObject({
      after: "2026-09-01",
      before: null,
      invalidDates: ["before:2026-02-30", "after:2026/09/01", "before:today"],
      keywords: [],
    });
  });

  it("解釈できない修飾子は語として残す", () => {
    expect(parseSearchQuery("from: has:audio is:foo during:week url:x")).toMatchObject({
      has: [],
      keywords: ["from:", "has:audio", "is:foo", "during:week", "url:x"],
    });
    expect(parseSearchQuery("from:@ in:#").keywords).toEqual(["from:@", "in:#"]);
  });

  it("空白だけなら条件なし", () => {
    expect(parseSearchQuery("  　 ")).toEqual(emptySearchQuery);
  });
});

describe("formatSearchQuery", () => {
  it("語のあとに修飾子を決まった順で並べる", () => {
    expect(
      formatSearchQuery({
        after: "2026-09-01",
        before: "2026-09-30",
        from: ["Alice Smith", "bob"],
        has: ["link"],
        in: ["dev"],
        invalidDates: ["after:yesterday"],
        is: ["pinned"],
        keywords: ["設計", "レビュー"],
      }),
    ).toBe(
      '設計 レビュー after:yesterday from:@"Alice Smith" from:@bob in:#dev has:link is:pinned after:2026-09-01 before:2026-09-30',
    );
  });

  it("解析した結果を戻すと同じ条件になる", () => {
    const raw = 'from:@"Alice Smith" in:#dev has:image before:2026-09-30 after:9/1 手順';
    expect(parseSearchQuery(formatSearchQuery(parseSearchQuery(raw)))).toEqual(
      parseSearchQuery(raw),
    );
  });
});

describe("hasSearchConditions", () => {
  it("語以外の条件があるか", () => {
    expect(hasSearchConditions(parseSearchQuery("release"))).toBe(false);
    expect(hasSearchConditions(parseSearchQuery("is:mention"))).toBe(true);
    expect(hasSearchConditions(parseSearchQuery("before:2026-01-01"))).toBe(true);
  });
});

describe("searchDateRange", () => {
  // ローカルタイムで組み立てるため実行環境のタイムゾーンに依存しない
  it("after / before は指定した日を含む", () => {
    expect(searchDateRange(parseSearchQuery("after:2026-09-01 before:2026-09-10"))).toEqual({
      after: new Date(2026, 8, 1),
      before: new Date(2026, 8, 11),
    });
    expect(searchDateRange(parseSearchQuery("before:2026-12-31")).before).toEqual(
      new Date(2027, 0, 1),
    );
  });

  it("条件がなければ範囲なし", () => {
    expect(searchDateRange(emptySearchQuery)).toEqual({ after: undefined, before: undefined });
  });
});
