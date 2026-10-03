import { atomWithStorage } from "jotai/utils";

// サイドバーのセクションの折りたたみ状態。キーはセクション名
export const collapsedSidebarSectionsAtom = atomWithStorage<Record<string, boolean>>(
  "sidebar-collapsed",
  {},
);

// 左サイドバーと右パネルの幅（px）。端末ごとに持つ
export const sidebarWidthRanges = {
  left: { defaultValue: 248, maxValue: 400, minValue: 200 },
  right: { defaultValue: 340, maxValue: 560, minValue: 280 },
};
export const sidebarWidthsAtom = atomWithStorage(
  "sidebar-widths",
  { left: sidebarWidthRanges.left.defaultValue, right: sidebarWidthRanges.right.defaultValue },
  undefined,
  { getOnInit: true },
);
