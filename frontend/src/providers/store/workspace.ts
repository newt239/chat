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

// ワークスペースを設定（ユーザー操作による切り替え。チャンネル選択は解除する）
export const setCurrentWorkspaceAtom = atom(null, (_get, set, workspaceId: string) => {
  set(workspaceStorageAtom, { currentWorkspaceId: workspaceId });
  set(currentChannelIdAtom, null);
});

// URL を情報源としてワークスペースを同期する。
// 表示中のチャンネルも URL から決まるため、こちらはチャンネル選択を解除しない。
export const syncCurrentWorkspaceAtom = atom(null, (get, set, workspaceId: string) => {
  if (get(workspaceStorageAtom).currentWorkspaceId !== workspaceId) {
    set(workspaceStorageAtom, { currentWorkspaceId: workspaceId });
  }
});

// チャンネルを設定
export const setCurrentChannelAtom = atom(null, (_get, set, channelId: string) => {
  set(currentChannelIdAtom, channelId);
});
