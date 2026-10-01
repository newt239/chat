import { describe, expect, test } from "vite-plus/test";

import { applySuggestion, findSuggestionQuery, rankByQuery } from "./suggestion";

describe("findSuggestionQuery", () => {
  test("行頭や空白の後の @ と # をカーソルまで検索語にする", () => {
    expect(findSuggestionQuery("@al", 3)).toEqual({ query: "al", start: 0, trigger: "@" });
    expect(findSuggestionQuery("see #dev/fr", 11)).toEqual({
      query: "dev/fr",
      start: 4,
      trigger: "#",
    });
    expect(findSuggestionQuery("hi @", 4)).toEqual({ query: "", start: 3, trigger: "@" });
    expect(findSuggestionQuery("@山田", 3)).toEqual({ query: "山田", start: 0, trigger: "@" });
  });

  test("単語の途中や空白を挟んだ後は候補を出さない", () => {
    expect(findSuggestionQuery("mail@example", 12)).toBeNull();
    expect(findSuggestionQuery("@alice hello", 12)).toBeNull();
  });
});

describe("applySuggestion", () => {
  test("検索語を置き換えて空白を足し、カーソルをその後ろに置く", () => {
    const text = "hi @al and more";
    const query = findSuggestionQuery(text, 6);
    expect(query && applySuggestion({ cursor: 6, query, text, value: "@Alice" })).toEqual({
      cursor: 10,
      text: "hi @Alice  and more",
    });
  });
});

describe("rankByQuery", () => {
  test("前方一致を先に並べ、一致しないものは除く", () => {
    const names = ["dev/backend", "backend", "design", "general"];
    expect(rankByQuery(names, "back", (name) => name)).toEqual(["backend", "dev/backend"]);
    expect(rankByQuery(names, "", (name) => name)).toEqual(names);
  });
});
