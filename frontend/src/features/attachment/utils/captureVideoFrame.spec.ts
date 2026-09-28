import { describe, expect, test } from "vite-plus/test";

import { fitThumbnail } from "./captureVideoFrame";

describe("fitThumbnail", () => {
  test("長辺が 960px を超えるときは比率を保って縮める", () => {
    expect(fitThumbnail(1920, 1080)).toEqual({ height: 540, width: 960 });
    expect(fitThumbnail(1080, 1920)).toEqual({ height: 960, width: 540 });
  });

  test("小さい動画はそのままの寸法にする", () => {
    expect(fitThumbnail(640, 360)).toEqual({ height: 360, width: 640 });
  });
});
