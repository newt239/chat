import { describe, expect, test } from "vite-plus/test";

import { ancestorPaths, lastSegment, parentPath, validateChannelPath } from "./channelPath";

describe("channelPath", () => {
  test("祖先・親・末尾を取り出す", () => {
    expect(ancestorPaths("dev/frontend/web")).toEqual(["dev", "dev/frontend"]);
    expect(parentPath("dev/frontend/web")).toBe("dev/frontend");
    expect(parentPath("dev")).toBeNull();
    expect(lastSegment("dev/frontend")).toBe("frontend");
  });

  test.each([
    ["", "required"],
    ["dev/", "segmentEmpty"],
    ["a/b/c/d/e", "tooDeep"],
    ["Dev", "invalid"],
    ["開発", "invalid"],
    [`dev/${"a".repeat(33)}`, "segmentTooLong"],
    ["general", "exists"],
    ["dev/front-end_2", null],
  ])("%s -> %s", (path, expected) => {
    expect(validateChannelPath(path, ["general"])).toBe(expected);
  });
});
