import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

// パネルの表示状態を管理する型定義
export type PanelView =
  | { type: "hidden" }
  | { type: "members" }
  | { type: "channel-info"; channelId?: string | null }
  | { type: "thread"; threadId: string }
  | { type: "user-profile"; userId: string }
  | { type: "search"; query: string; filter: "all" | "messages" | "channels" | "users" }
  | { type: "bookmarks" };

// レイアウトの状態を管理する型定義
export type LayoutState = {
  // 左サイドパネルの表示状態
  leftSidePanelVisible: boolean;
  // 右サイドパネルの表示状態と内容
  rightSidePanelView: PanelView;
  // モバイルビューの状態
  isMobile: boolean;
  // モバイルで表示中のパネル（left, right, none）
  mobileActivePanel: "left" | "right" | "none";
};

const defaultLayoutState: LayoutState = {
  leftSidePanelVisible: true,
  rightSidePanelView: { type: "hidden" },
  isMobile: false,
  mobileActivePanel: "none",
};

// レイアウト状態のAtom
export const layoutStateAtom = atomWithStorage<LayoutState>(
  "ui-storage:layoutState",
  defaultLayoutState
);

// 個別の状態を取得するAtom
export const leftSidePanelVisibleAtom = atom((get) => get(layoutStateAtom).leftSidePanelVisible);

export const rightSidePanelViewAtom = atom((get) => get(layoutStateAtom).rightSidePanelView);

export const isMobileAtom = atom((get) => get(layoutStateAtom).isMobile);

export const mobileActivePanelAtom = atom((get) => get(layoutStateAtom).mobileActivePanel);

// パネル表示状態の比較関数
const isSamePanelView = (first: PanelView, second: PanelView): boolean => {
  if (first.type !== second.type) {
    return false;
  }

  switch (first.type) {
    case "hidden":
    case "members":
    case "bookmarks":
      return true;
    case "channel-info":
      return second.type === "channel-info" && first.channelId === second.channelId;
    case "thread":
      return second.type === "thread" && first.threadId === second.threadId;
    case "user-profile":
      return second.type === "user-profile" && first.userId === second.userId;
    case "search":
      return (
        second.type === "search" && first.query === second.query && first.filter === second.filter
      );
  }
};

// 左サイドパネルの表示/非表示を切り替える
export const toggleLeftSidePanelAtom = atom(null, (get, set) => {
  const current = get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    leftSidePanelVisible: !current.leftSidePanelVisible,
  });
});

// 左サイドパネルを表示する
export const showLeftSidePanelAtom = atom(null, (_get, set) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    leftSidePanelVisible: true,
  });
});

// 左サイドパネルを非表示にする
export const hideLeftSidePanelAtom = atom(null, (_get, set) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    leftSidePanelVisible: false,
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

// 右サイドパネルの表示内容を切り替える
export const toggleRightSidePanelViewAtom = atom(null, (get, set, view: PanelView) => {
  const current = get(layoutStateAtom);
  if (isSamePanelView(current.rightSidePanelView, view)) {
    set(layoutStateAtom, {
      ...current,
      rightSidePanelView: { type: "hidden" },
    });
    return;
  }
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

// モバイルビューの状態を設定する
export const setIsMobileAtom = atom(null, (_get, set, isMobile: boolean) => {
  const current = _get(layoutStateAtom);
  set(layoutStateAtom, {
    ...current,
    isMobile,
    // モバイルでない場合はアクティブパネルをリセット
    mobileActivePanel: isMobile ? current.mobileActivePanel : "none",
  });
});

// モバイルでアクティブなパネルを設定する
export const setMobileActivePanelAtom = atom(
  null,
  (_get, set, panel: "left" | "right" | "none") => {
    const current = _get(layoutStateAtom);
    set(layoutStateAtom, {
      ...current,
      mobileActivePanel: panel,
    });
  }
);

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

// レガシー互換性のためのエクスポート（既存コードとの互換性を保つ）
export const rightSidebarViewAtom = rightSidePanelViewAtom;
export const setRightSidebarViewAtom = setRightSidePanelViewAtom;
export const toggleRightSidebarViewAtom = toggleRightSidePanelViewAtom;
export const closeRightSidebarAtom = closeRightSidePanelAtom;
