import { useEffect, useEffectEvent } from "react";

import { useCanGoBack, useMatch, useNavigate, useRouter } from "@tanstack/react-router";

import { useNextUnreadChannel } from "#/features/channel/hooks/useNextUnreadChannel";

// ⌘K（Ctrl+K）で検索、⌘,（Ctrl+,）で設定を開く。設定ページで押すと元の画面へ戻る。⌥⇧↓（Alt+Shift+↓）で次の未読へ移る
export const useGlobalShortcuts = (workspaceId: string) => {
  const navigate = useNavigate();
  const router = useRouter();
  const canGoBack = useCanGoBack();
  const { goNext } = useNextUnreadChannel(workspaceId);
  const isSettingsOpen =
    useMatch({ from: "/app/$workspaceId/settings/{-$section}", shouldThrow: false }) !== undefined;

  const handleKeyDown = useEffectEvent((event: KeyboardEvent) => {
    if (
      event.altKey &&
      event.shiftKey &&
      !event.metaKey &&
      !event.ctrlKey &&
      event.key === "ArrowDown"
    ) {
      event.preventDefault();
      goNext();
      return;
    }
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
          to: "/app/$workspaceId/settings/{-$section}",
        });
      } else if (canGoBack) {
        router.history.back();
      } else {
        void navigate({ params: { workspaceId }, to: "/app/$workspaceId" });
      }
    }
  });

  useEffect(() => {
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);
};
