import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

// パネルの表示状態を管理する型定義
export type PanelView =
  | { type: "hidden" }
  | { type: "channel-members"; channelId: string }
  | { type: "channel-info"; channelId?: string | null }
  | { type: "thread"; threadId: string }
  | { type: "pins"; channelId: string }
  | { type: "user-profile"; userId: string }
  | { type: "search"; query: string; filter: "all" | "messages" | "channels" | "users" }
  | { type: "bookmarks" }
  | { type: "notifications" };

// レイアウトの状態を管理する型定義
type LayoutState = {
  // 左サイドパネルの表示状態
  leftSidePanelVisible: boolean;
  // 右サイドパネルの表示状態と内容
  rightSidePanelView: PanelView;
  // モバイルで表示中のパネル（left, right, none）
  mobileActivePanel: "left" | "right" | "none";
  // 現在のルートがチャンネルページかどうか
  isChannelPage: boolean;
  // チャンネル別のピン件数
  pinsCountByChannel: Record<string, number>;
};

const defaultLayoutState: LayoutState = {
  leftSidePanelVisible: true,
  rightSidePanelView: { type: "hidden" },
  mobileActivePanel: "none",
  isChannelPage: false,
  pinsCountByChannel: {},
};

// レイアウト状態のAtom
const layoutStateAtom = atomWithStorage<LayoutState>("ui-storage:layoutState", defaultLayoutState);

// 個別の状態を取得するAtom
export const leftSidePanelVisibleAtom = atom((get) => get(layoutStateAtom).leftSidePanelVisible);

export const rightSidePanelViewAtom = atom((get) => get(layoutStateAtom).rightSidePanelView);

export const mobileActivePanelAtom = atom((get) => get(layoutStateAtom).mobileActivePanel);

export const isChannelPageAtom = atom((get) => get(layoutStateAtom).isChannelPage);

// チャンネル別ピン件数の取得
export const pinsCountByChannelAtom = atom((get) => get(layoutStateAtom).pinsCountByChannel);

// チャンネルのピン件数を設定
export const setChannelPinsCountAtom = atom(
  null,
  (_get, set, payload: { channelId: string; count: number }) => {
    const current = _get(layoutStateAtom);
    set(layoutStateAtom, {
      ...current,
      pinsCountByChannel: {
        ...current.pinsCountByChannel,
        [payload.channelId]: payload.count,
      },
    });
  }
);

// チャンネルのピン件数をインクリメント/デクリメント
export const addChannelPinsDeltaAtom = atom(
  null,
  (_get, set, payload: { channelId: string; delta: number }) => {
    const current = _get(layoutStateAtom);
    const prev = current.pinsCountByChannel[payload.channelId] ?? 0;
    const next = Math.max(0, prev + payload.delta);
    set(layoutStateAtom, {
      ...current,
      pinsCountByChannel: {
        ...current.pinsCountByChannel,
        [payload.channelId]: next,
      },
    });
  }
);

// 左サイドパネルを表示する
export const showLeftSidePanelAtom = atom(null, (_get, set) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    leftSidePanelVisible: true,
  });
});

// 右サイドパネルの表示内容を設定する
export const setRightSidePanelViewAtom = atom(null, (_get, set, view: PanelView) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    rightSidePanelView: view,
  });
});

// 右サイドパネルを閉じる
export const closeRightSidePanelAtom = atom(null, (_get, set) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    rightSidePanelView: { type: "hidden" },
  });
});

// モバイルで左パネルを表示する
export const showMobileLeftPanelAtom = atom(null, (_get, set) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    mobileActivePanel: "left",
  });
});

// モバイルで右パネルを表示する
export const showMobileRightPanelAtom = atom(null, (_get, set) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    mobileActivePanel: "right",
  });
});

// モバイルでパネルを閉じる
export const hideMobilePanelsAtom = atom(null, (_get, set) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    mobileActivePanel: "none",
  });
});

// チャンネルページの状態を設定する
export const setIsChannelPageAtom = atom(null, (_get, set, isChannelPage: boolean) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    isChannelPage,
  });
});
