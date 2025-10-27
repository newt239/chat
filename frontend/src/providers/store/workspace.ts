import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

import { navigateToWorkspace, navigateToChannel } from "@/lib/navigation";

type WorkspaceStorage = {
  currentWorkspaceId: string | null;
};

// ワークスペースIDをストレージに保存
export const workspaceStorageAtom = atomWithStorage<WorkspaceStorage>(
  "workspace-storage",
  {
    currentWorkspaceId: null,
  },
  undefined,
  { getOnInit: true }
);

// 現在のワークスペースID
export const currentWorkspaceIdAtom = atom<string | null>(
  (get) => get(workspaceStorageAtom).currentWorkspaceId
);

// 現在のチャンネルID（メモリのみ、永続化しない）
export const currentChannelIdAtom = atom<string | null>(null);

// ワークスペースを設定
export const setCurrentWorkspaceAtom = atom(null, (_get, set, workspaceId: string) => {
  set(workspaceStorageAtom, { currentWorkspaceId: workspaceId });
  set(currentChannelIdAtom, null);
});

// チャンネルを設定
export const setCurrentChannelAtom = atom(null, (_get, set, channelId: string) => {
  set(currentChannelIdAtom, channelId);
});

// ワークスペースを設定してナビゲート
export const setCurrentWorkspaceAndNavigateAtom = atom(null, (_get, set, workspaceId: string) => {
  set(workspaceStorageAtom, { currentWorkspaceId: workspaceId });
  set(currentChannelIdAtom, null);
  navigateToWorkspace(workspaceId);
});

// チャンネルを設定してナビゲート
export const setCurrentChannelAndNavigateAtom = atom(
  null,
  (_get, set, args: { workspaceId: string; channelId: string }) => {
    const { workspaceId, channelId } = args;
    set(workspaceStorageAtom, { currentWorkspaceId: workspaceId });
    set(currentChannelIdAtom, channelId);
    navigateToChannel(workspaceId, channelId);
  }
);
