import { describe, expect, test } from "vite-plus/test";

import { transitions } from "./motion";

describe("transitions", () => {
  test("時間のトークンは秒に変換する", () => {
    expect(transitions.fast).toStrictEqual({ duration: 0.09, ease: [0.2, 0, 0, 1] });
  });

  test("スプリングのトークンはそのまま渡す", () => {
    expect(transitions.sheet).toStrictEqual({
      damping: 44,
      mass: 1,
      stiffness: 560,
      type: "spring",
    });
  });
});
