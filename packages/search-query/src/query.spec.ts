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
      parseSearchQuery("リリース from:@alice in:#dev/frontend has:image during:week 手順"),
    ).toEqual({
      ...emptySearchQuery,
      during: "week",
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

  it("同じ条件は重ねず、期間は後のものを使う", () => {
    expect(
      parseSearchQuery("has:file has:file is:thread is:mention during:today during:month"),
    ).toMatchObject({ during: "month", has: ["file"], is: ["thread", "mention"] });
  });

  it("クォートで空白を含む名前を指定できる", () => {
    expect(parseSearchQuery('from:@"Alice Smith" "foo bar"')).toMatchObject({
      from: ["Alice Smith"],
      keywords: ['"foo bar"'],
    });
  });

  it("日付を受け付け、存在しない日付は語として残す", () => {
    expect(parseSearchQuery("after:2026-09-01 before:2026-02-30")).toMatchObject({
      after: "2026-09-01",
      before: null,
      keywords: ["before:2026-02-30"],
    });
  });

  it("解釈できない修飾子は語として残す", () => {
    expect(parseSearchQuery("from: has:audio is:foo during:year url:x")).toMatchObject({
      has: [],
      keywords: ["from:", "has:audio", "is:foo", "during:year", "url:x"],
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
        during: "today",
        from: ["Alice Smith", "bob"],
        has: ["link"],
        in: ["dev"],
        is: ["pinned"],
        keywords: ["設計", "レビュー"],
      }),
    ).toBe(
      '設計 レビュー from:@"Alice Smith" from:@bob in:#dev has:link is:pinned during:today after:2026-09-01 before:2026-09-30',
    );
  });

  it("解析した結果を戻すと同じ条件になる", () => {
    const raw = 'from:@"Alice Smith" in:#dev has:image during:week 手順';
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
  const now = new Date(2026, 8, 29, 15, 30);

  it("during は今日を含む日数の 0 時から", () => {
    expect(searchDateRange(parseSearchQuery("during:today"), now)).toEqual({
      after: new Date(2026, 8, 29),
      before: undefined,
    });
    expect(searchDateRange(parseSearchQuery("during:week"), now).after).toEqual(
      new Date(2026, 8, 23),
    );
    expect(searchDateRange(parseSearchQuery("during:month"), now).after).toEqual(
      new Date(2026, 7, 31),
    );
  });

  it("after / before は指定した日を含まない", () => {
    expect(searchDateRange(parseSearchQuery("after:2026-09-01 before:2026-09-10"), now)).toEqual({
      after: new Date(2026, 8, 2),
      before: new Date(2026, 8, 10),
    });
  });

  it("重なる条件は狭い方を使う", () => {
    expect(
      searchDateRange(parseSearchQuery("during:month after:2026-09-20 before:2026-09-25"), now),
    ).toEqual({ after: new Date(2026, 8, 21), before: new Date(2026, 8, 25) });
  });

  it("条件がなければ範囲なし", () => {
    expect(searchDateRange(emptySearchQuery, now)).toEqual({
      after: undefined,
      before: undefined,
    });
  });
});
