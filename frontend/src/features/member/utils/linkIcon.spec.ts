import { IconBrandGithub, IconBrandX, IconBrandYoutube, IconLink } from "@tabler/icons-react";
import { describe, expect, test } from "vite-plus/test";

import { displayUrl, linkIconOf } from "./linkIcon";

describe("linkIconOf", () => {
  test("主要なサイトはそのアイコン、それ以外は汎用のアイコンにする", () => {
    expect(linkIconOf("https://x.com/newt239")).toBe(IconBrandX);
    expect(linkIconOf("https://www.github.com/newt239")).toBe(IconBrandGithub);
    expect(linkIconOf("https://m.youtube.com/@newt")).toBe(IconBrandYoutube);
    expect(linkIconOf("https://notgithub.com")).toBe(IconLink);
    expect(linkIconOf("not a url")).toBe(IconLink);
  });
});

describe("displayUrl", () => {
  test("スキーム・www.・末尾のスラッシュを除く", () => {
    expect(displayUrl("https://github.com/newt239/")).toBe("github.com/newt239");
    expect(displayUrl("https://www.youtube.com/@newt")).toBe("youtube.com/@newt");
  });
});
