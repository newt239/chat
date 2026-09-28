import { describe, expect, test } from "vite-plus/test";

import { imageBox } from "./imageBox";

describe("imageBox", () => {
  test("大きい画像は比率を保って 400×300 に収める", () => {
    expect(imageBox(1600, 900)).toEqual({ crop: null, height: 225, width: 400 });
    expect(imageBox(900, 1600)).toEqual({ crop: null, height: 300, width: 169 });
  });

  test("小さい画像は拡大しない", () => {
    expect(imageBox(200, 150)).toEqual({ crop: null, height: 150, width: 200 });
  });

  test("極端に縦長・横長の画像は最小サイズまで広げて切り取る", () => {
    expect(imageBox(400, 4000)).toEqual({ crop: "tall", height: 300, width: 140 });
    expect(imageBox(4000, 200)).toEqual({ crop: "wide", height: 90, width: 400 });
  });

  test("寸法が分からなければ 4:3 の枠にする", () => {
    expect(imageBox(undefined, undefined)).toEqual({ crop: null, height: 300, width: 400 });
  });
});
