import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

const pinsCountAtom = atom<Record<string, number>>({});

export const pinsCountByChannelAtom = atom((get) => get(pinsCountAtom));

export const setChannelPinsCountAtom = atom(
  null,
  (get, set, payload: { channelId: string; count: number }) => {
    set(pinsCountAtom, { ...get(pinsCountAtom), [payload.channelId]: payload.count });
  },
);

export const addChannelPinsDeltaAtom = atom(
  null,
  (get, set, payload: { channelId: string; delta: number }) => {
    const current = get(pinsCountAtom);
    const next = Math.max(0, (current[payload.channelId] ?? 0) + payload.delta);
    set(pinsCountAtom, { ...current, [payload.channelId]: next });
  },
);

// モバイルで最後に開いたボトムタブ。チャンネルなどを開いている間も下に残す
export type MobileTab = "home" | "dms" | "activity" | "me";
export const mobileTabAtom = atom<MobileTab>("home");

// サイドバーのセクションの折りたたみ状態。キーはセクション名
export const collapsedSidebarSectionsAtom = atomWithStorage<Record<string, boolean>>(
  "sidebar-collapsed",
  {},
);

// サイドバーのチャンネルツリーで折りたたんだ親。キーはチャンネル ID。端末ごとに持つ
export const collapsedChannelsAtom = atomWithStorage<Record<string, boolean>>(
  "channel-tree-collapsed",
  {},
);

// 親チャンネルで「下階層を含む」をオフにしたチャンネル。既定はオン。端末ごとに持つ
export const excludedDescendantsAtom = atomWithStorage<Record<string, boolean>>(
  "channel-descendants-excluded",
  {},
);

// チャンネルごとの閲覧中のユーザー。WebSocket の channelViewers で置き換える
export const channelViewersAtom = atom<Record<string, string[]>>({});

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
