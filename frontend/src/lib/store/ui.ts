import { create } from "zustand";
import { persist } from "zustand/middleware";

interface UIState {
  isMemberPanelOpen: boolean;
  toggleMemberPanel: () => void;
  setMemberPanelOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      isMemberPanelOpen: true,
      toggleMemberPanel: () =>
        set((state) => ({ isMemberPanelOpen: !state.isMemberPanelOpen })),
      setMemberPanelOpen: (open) => set({ isMemberPanelOpen: open }),
    }),
    {
      name: "ui-storage",
    }
  )
);
