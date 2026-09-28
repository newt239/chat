import { useEffect } from "react";

import { useNavigate } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { settingsSectionAtom } from "#/providers/store/ui";

/** ⌘K（Ctrl+K）で検索、⌘,（Ctrl+,）で設定を開く */
export const useGlobalShortcuts = (workspaceId: string) => {
  const navigate = useNavigate();
  const setSettingsSection = useSetAtom(settingsSectionAtom);

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
        setSettingsSection((current) => (current === null ? "theme" : null));
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [navigate, setSettingsSection, workspaceId]);
};
