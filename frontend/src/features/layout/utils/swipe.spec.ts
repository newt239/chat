import { describe, expect, test } from "vite-plus/test";

import { isSwipeBlocked } from "./swipe";

const setup = (html: string) => {
  const root = document.createElement("div");
  root.innerHTML = html;
  document.body.append(root);
  return root;
};

describe("isSwipeBlocked", () => {
  test("入力欄の中から始まった操作はスワイプにしない", () => {
    const root = setup("<label><textarea></textarea></label><p>本文</p>");
    expect(isSwipeBlocked(root.querySelector("textarea") ?? root, root)).toBe(true);
    expect(isSwipeBlocked(root.querySelector("p") ?? root, root)).toBe(false);
  });

  test("横にスクロールできる要素の中から始まった操作はスワイプにしない", () => {
    const root = setup('<div style="overflow-x: auto"><span>code</span></div>');
    const scroller = root.querySelector("div") ?? root;
    Object.defineProperties(scroller, {
      clientWidth: { value: 100 },
      scrollWidth: { value: 300 },
    });
    expect(isSwipeBlocked(root.querySelector("span") ?? root, root)).toBe(true);
  });
});
