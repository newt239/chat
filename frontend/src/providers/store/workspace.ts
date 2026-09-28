import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

type WorkspaceStorage = {
  currentWorkspaceId: string | null;
};

// ワークスペースIDをストレージに保存
const workspaceStorageAtom = atomWithStorage<WorkspaceStorage>(
  "workspace-storage",
  {
    currentWorkspaceId: null,
  },
  undefined,
  { getOnInit: true },
);

// 現在のワークスペースID
export const currentWorkspaceIdAtom = atom<string | null>(
  (get) => get(workspaceStorageAtom).currentWorkspaceId,
);

// 現在のチャンネルID（メモリのみ、永続化しない）
export const currentChannelIdAtom = atom<string | null>(null);

// URL から同期する用。チャンネルも URL で決まるため選択を解除しない
export const syncCurrentWorkspaceAtom = atom(null, (get, set, workspaceId: string) => {
  if (get(workspaceStorageAtom).currentWorkspaceId !== workspaceId) {
    set(workspaceStorageAtom, { currentWorkspaceId: workspaceId });
  }
});

// チャンネルを設定
export const setCurrentChannelAtom = atom(null, (_get, set, channelId: string) => {
  set(currentChannelIdAtom, channelId);
});
