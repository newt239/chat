import { describe, expect, test } from "vite-plus/test";

import { toPlainText } from "./plainText";

describe("toPlainText", () => {
  test("コードブロックと記号を除き、空白をまとめる", () => {
    expect(toPlainText("手順です。\n\n```ts\nconst a = 1;\n```\n\n- **確認**\n> 引用")).toBe(
      "手順です。 確認 引用",
    );
  });
});
