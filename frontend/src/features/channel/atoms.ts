import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

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
