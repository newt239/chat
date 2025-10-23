import { create } from "zustand";

interface WorkspaceState {
  currentWorkspaceId: string | null;
  currentChannelId: string | null;
  setCurrentWorkspace: (workspaceId: string) => void;
  setCurrentChannel: (channelId: string) => void;
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
  currentWorkspaceId: null,
  currentChannelId: null,
  setCurrentWorkspace: (workspaceId) => set({ currentWorkspaceId: workspaceId }),
  setCurrentChannel: (channelId) => set({ currentChannelId: channelId }),
}));
