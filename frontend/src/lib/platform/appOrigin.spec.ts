import { expect, test } from "vite-plus/test";

import { toShareUrl } from "#/lib/platform/appOrigin";

test("相対パスを origin 付きの絶対 URL にする", () => {
  expect(toShareUrl("/app/w1/c1?message=m1")).toBe(
    `${globalThis.location.origin}/app/w1/c1?message=m1`,
  );
});
