import { describe, expect, test } from "vite-plus/test";

import { continueList, detectActiveFormats, insertEmoji, toggleFormat } from "./format";

const at = (position: number) => ({ end: position, start: position });

describe("toggleFormat", () => {
  test("選択範囲を記号で囲み、選択を内側の文字に保つ", () => {
    expect(toggleFormat("hello world", { end: 11, start: 6 }, "bold")).toEqual({
      selection: { end: 13, start: 8 },
      text: "hello **world**",
    });
  });

  test("太字の内側でもう一度押すと記号を外す", () => {
    expect(toggleFormat("hello **world**", { end: 13, start: 8 }, "bold")).toEqual({
      selection: { end: 11, start: 6 },
      text: "hello world",
    });
    expect(toggleFormat("a **bc** d", at(5), "bold").text).toBe("a bc d");
  });

  test("記号ごと選択していても外す", () => {
    expect(toggleFormat("~~x~~", { end: 5, start: 0 }, "strikethrough")).toEqual({
      selection: { end: 1, start: 0 },
      text: "x",
    });
  });

  test("アスタリスクの斜体も外せる。太字の記号は斜体と見なさない", () => {
    expect(toggleFormat("*x*", at(2), "italic").text).toBe("x");
    expect(toggleFormat("**x**", at(3), "italic").text).toBe("**x__**");
  });

  test("リンクは url を選択した状態で挿入し、リンクの中で押すと文字だけ残す", () => {
    expect(toggleFormat("", at(0), "link")).toEqual({
      selection: { end: 6, start: 3 },
      text: "[](url)",
    });
    expect(toggleFormat("see [docs](https://a.b) now", at(6), "link").text).toBe("see docs now");
  });

  test("行頭の記号は選択した各行に付け、番号は順に振る", () => {
    expect(toggleFormat("a\nb\n\nc", { end: 6, start: 0 }, "orderedList").text).toBe(
      "1. a\n2. b\n\n3. c",
    );
    expect(toggleFormat("x\nfoo", at(4), "quote")).toEqual({
      selection: at(6),
      text: "x\n> foo",
    });
  });

  test("すべての行に付いていれば外し、箇条書きと番号付きは置き換える", () => {
    expect(toggleFormat("- a\n* b", { end: 7, start: 0 }, "list").text).toBe("a\nb");
    expect(toggleFormat("- a\n- b", { end: 7, start: 0 }, "orderedList").text).toBe("1. a\n2. b");
  });
});

describe("detectActiveFormats", () => {
  test("カーソルが太字の内側にあれば bold が有効", () => {
    const text = "a **bold** b";
    expect(detectActiveFormats(text, at(5)).bold).toBe(true);
    expect(detectActiveFormats(text, at(11)).bold).toBe(false);
  });

  test("隣り合う別の対の間は書式の内側と見なさない", () => {
    expect(detectActiveFormats("*a* b *c*", at(4)).italic).toBe(false);
    expect(detectActiveFormats("*a* b *c*", at(7)).italic).toBe(true);
  });

  test("行頭の記号から見出しやリストを判定する。* や + も箇条書き", () => {
    expect(detectActiveFormats("first\n- item", at(12))).toMatchObject({
      heading: false,
      list: true,
    });
    expect(detectActiveFormats("* item", at(6)).list).toBe(true);
    expect(detectActiveFormats("+ item", at(6)).list).toBe(true);
  });
});

describe("continueList", () => {
  test("箇条書きの記号を次の行に引き継ぐ", () => {
    expect(continueList("* a", at(3))).toEqual({ selection: at(6), text: "* a\n* " });
    expect(continueList("  - a", at(5))?.text).toBe("  - a\n  - ");
  });

  test("番号付きは次の番号、タスクは未完了の項目にする", () => {
    expect(continueList("1. a", at(4))?.text).toBe("1. a\n2. ");
    expect(continueList("- [x] done", at(10))?.text).toBe("- [x] done\n- [ ] ");
  });

  test("空の項目で改行するとリストを抜ける", () => {
    expect(continueList("- a\n- ", at(6))).toEqual({ selection: at(4), text: "- a\n" });
  });

  test("リストでない行や範囲選択では何もしない", () => {
    expect(continueList("plain", at(5))).toBeNull();
    expect(continueList("- a", { end: 3, start: 2 })).toBeNull();
  });
});

describe("insertEmoji", () => {
  test("選択範囲を絵文字で置き換え、カーソルを直後に置く", () => {
    expect(insertEmoji("いいね！", { end: 4, start: 3 }, "🎉")).toEqual({
      cursor: 5,
      text: "いいね🎉",
    });
  });

  test("カスタム絵文字は英数字と接するときだけ空白を挟む", () => {
    expect(insertEmoji("ok", { end: 2, start: 2 }, ":party:").text).toBe("ok :party:");
    expect(insertEmoji("abc", { end: 1, start: 1 }, ":party:").text).toBe("a :party: bc");
    expect(insertEmoji("了解", { end: 2, start: 2 }, ":party:").text).toBe("了解:party:");
  });
});
