import { useEffect } from "react";

import { useNavigate } from "@tanstack/react-router";

import { closeDialog, openDialog } from "../utils/overlaySearch";
import { workspaceRoute } from "../utils/workspaceRoute";

/** ⌘K（Ctrl+K）で検索、⌘,（Ctrl+,）で設定を開く */
export const useGlobalShortcuts = (workspaceId: string) => {
  const navigate = useNavigate();
  const isSettingsOpen = workspaceRoute.useSearch({
    select: (search) => search.settings !== undefined,
  });

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!(event.metaKey || event.ctrlKey) || event.altKey || event.shiftKey) {
        return;
      }
      if (event.key.toLowerCase() === "k") {
        event.preventDefault();
        void navigate({ params: { workspaceId }, to: "/app/$workspaceId/search" });
      }
      if (event.key === ",") {
        event.preventDefault();
        void navigate({
          search: isSettingsOpen ? closeDialog : openDialog({ settings: "theme" }),
          to: ".",
        });
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [navigate, isSettingsOpen, workspaceId]);
};
