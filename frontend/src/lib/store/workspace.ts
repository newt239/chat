import { create } from "zustand";
import { persist } from "zustand/middleware";

import { navigateToWorkspace, navigateToChannel } from "../navigation";

interface WorkspaceState {
  currentWorkspaceId: string | null;
  currentChannelId: string | null;
  setCurrentWorkspace: (workspaceId: string) => void;
  setCurrentChannel: (channelId: string) => void;
  setCurrentWorkspaceAndNavigate: (workspaceId: string) => void;
  setCurrentChannelAndNavigate: (workspaceId: string, channelId: string) => void;
}

export const useWorkspaceStore = create<WorkspaceState>()(
  persist(
    (set) => ({
      currentWorkspaceId: null,
      currentChannelId: null,
      setCurrentWorkspace: (workspaceId) =>
        set({
          currentWorkspaceId: workspaceId,
          currentChannelId: null,
        }),
      setCurrentChannel: (channelId) => set({ currentChannelId: channelId }),
      setCurrentWorkspaceAndNavigate: (workspaceId) => {
        set({
          currentWorkspaceId: workspaceId,
          currentChannelId: null,
        });
        navigateToWorkspace(workspaceId);
      },
      setCurrentChannelAndNavigate: (workspaceId, channelId) => {
        set({
          currentWorkspaceId: workspaceId,
          currentChannelId: channelId,
        });
        navigateToChannel(workspaceId, channelId);
      },
    }),
    {
      name: "workspace-storage",
      partialize: (state) => ({ currentWorkspaceId: state.currentWorkspaceId }),
    }
  )
);
