import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

import { navigateTo } from "#/lib/navigation";

import { currentChannelIdAtom, currentWorkspaceIdAtom } from "./workspace";

// 右パネルの内容。スレッドだけは URL（/thread/$messageId）で表す
export type PanelView =
  | { type: "hidden" }
  | { type: "channel-members"; channelId: string }
  | { type: "channel-info"; channelId?: string | null }
  | { type: "thread"; threadId: string }
  | { type: "pins"; channelId: string }
  | { type: "user-profile"; userId: string };

export type RightPanelView = Exclude<PanelView, { type: "thread" }>;

const rightPanelAtom = atom<RightPanelView>({ type: "hidden" });

type ThreadRouteParams = { workspaceId: string; channelId: string; messageId: string };

// 開いているスレッドのルート。スレッドのルートの onEnter / onLeave が書き込む
export const openThreadRouteAtom = atom<ThreadRouteParams | null>(null);

export const rightSidePanelViewAtom = atom((get) => get(rightPanelAtom));

// スレッドはそのチャンネルのスレッド URL へ遷移する。それ以外はスレッドを閉じてから表示する
export const setRightSidePanelViewAtom = atom(null, (get, set, view: PanelView) => {
  if (view.type === "thread") {
    const workspaceId = get(currentWorkspaceIdAtom);
    const channelId = get(currentChannelIdAtom);
    set(rightPanelAtom, { type: "hidden" });
    if (workspaceId !== null && channelId !== null) {
      navigateTo({
        params: { channelId, messageId: view.threadId, workspaceId },
        to: "/app/$workspaceId/$channelId/thread/$messageId",
      });
    }
    return;
  }
  set(rightPanelAtom, view);
  const thread = get(openThreadRouteAtom);
  if (view.type !== "hidden" && thread !== null) {
    navigateTo({
      params: { channelId: thread.channelId, workspaceId: thread.workspaceId },
      to: "/app/$workspaceId/$channelId",
    });
  }
});

export const closeRightSidePanelAtom = atom(null, (_get, set) => {
  set(rightPanelAtom, { type: "hidden" });
});

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

export const settingsSections = [
  "account",
  "notifications",
  "theme",
  "display",
  "shortcuts",
] as const;
export type SettingsSection = (typeof settingsSections)[number];

// 開いている設定の項目。null なら閉じている
export const settingsSectionAtom = atom<SettingsSection | null>(null);

// モバイルで最後に開いたボトムタブ。チャンネルなどを開いている間も下に残す
export type MobileTab = "home" | "dms" | "activity" | "me";
export const mobileTabAtom = atom<MobileTab>("home");

// サイドバーのセクションの折りたたみ状態。キーはセクション名
export const collapsedSidebarSectionsAtom = atomWithStorage<Record<string, boolean>>(
  "sidebar-collapsed",
  {},
);
