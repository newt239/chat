import { describe, expect, test } from "vite-plus/test";

import { CUSTOM_EMOJI_IN_TEXT, customEmojiName } from "./customEmoji";

const names = (text: string) => [...text.matchAll(CUSTOM_EMOJI_IN_TEXT)].map((m) => m.groups?.name);

describe("customEmoji", () => {
  test("本文の :name: を拾う。続けて書いたものも別々に拾う", () => {
    expect(names("やった :party: :tada::ok_hand:")).toEqual(["party", "tada", "ok_hand"]);
  });

  test("時刻や英数字に挟まれたコロンは拾わない", () => {
    expect(names("10:30:45 a:b:c :Party:")).toEqual([]);
  });

  test("リアクションの値が :name: のときだけ名前を返す", () => {
    expect(customEmojiName(":party:")).toBe("party");
    expect(customEmojiName("🎉")).toBeNull();
    expect(customEmojiName(":party: ")).toBeNull();
  });
});
