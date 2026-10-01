import { describe, expect, test } from "vite-plus/test";

import { findCommand, unescapeCommand } from "./commands";

describe("findCommand", () => {
  test("先頭の既知のコマンドだけを見つける", () => {
    expect(findCommand("/remind me 資料 10分後")).toBe("remind");
    expect(findCommand("/remind")).toBe("remind");
    expect(findCommand("/unknown x")).toBeNull();
    expect(findCommand("//remind me")).toBeNull();
    expect(findCommand("text /remind")).toBeNull();
  });
});

describe("unescapeCommand", () => {
  test("「//」で始まる入力は先頭の「/」を 1 つ外す", () => {
    expect(unescapeCommand("//remind はコマンドです")).toBe("/remind はコマンドです");
    expect(unescapeCommand("/path")).toBe("/path");
  });
});
