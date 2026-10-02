const blockingSelector = "input, textarea, [contenteditable], [role=slider], [data-swipe-ignore]";

// 入力欄やスライダー、横にスクロールできる要素の上で始まった操作は画面のスワイプにしない
export const isSwipeBlocked = (target: Element, root: Element) => {
  if (target.closest(blockingSelector) !== null) {
    return true;
  }
  for (
    let element: Element | null = target;
    element && element !== root;
    element = element.parentElement
  ) {
    const { overflowX } = getComputedStyle(element);
    if (
      (overflowX === "auto" || overflowX === "scroll") &&
      element.scrollWidth > element.clientWidth
    ) {
      return true;
    }
  }
  return false;
};
