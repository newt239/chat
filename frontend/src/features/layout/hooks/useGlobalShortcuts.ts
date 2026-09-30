import { useEffect } from "react";

import { useCanGoBack, useMatch, useNavigate, useRouter } from "@tanstack/react-router";

/** ⌘K（Ctrl+K）で検索、⌘,（Ctrl+,）で設定を開く。設定ページで押すと元の画面へ戻る */
export const useGlobalShortcuts = (workspaceId: string) => {
  const navigate = useNavigate();
  const router = useRouter();
  const canGoBack = useCanGoBack();
  const isSettingsOpen =
    useMatch({ from: "/app/$workspaceId/settings/$section", shouldThrow: false }) !== undefined;

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
        if (!isSettingsOpen) {
          void navigate({
            params: { section: "theme", workspaceId },
            to: "/app/$workspaceId/settings/$section",
          });
        } else if (canGoBack) {
          router.history.back();
        } else {
          void navigate({ params: { workspaceId }, to: "/app/$workspaceId" });
        }
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [navigate, router, canGoBack, isSettingsOpen, workspaceId]);
};
