import { useEffect } from "react";

import { Outlet, useCanGoBack, useMatches, useNavigate, useRouter } from "@tanstack/react-router";
import { useAtom } from "jotai";

import { DMsPage } from "#/features/dm/components/DMsPage";
import { useVisualViewport } from "#/features/layout/hooks/useVisualViewport";
import { MiniPlayer } from "#/features/player/components/MiniPlayer";
import { mobileTabAtom } from "#/providers/store/ui";

import { useRightPanel } from "../hooks/useRightPanel";
import { ActivityPage } from "./ActivityPage";
import { BackButton } from "./BackButton";
import { MePage } from "./MePage";
import { MobileHome } from "./MobileHome";
import { MobileStackLayer } from "./MobileStackLayer";
import { MobileTabBar } from "./MobileTabBar";

import type { MobileTab } from "#/providers/store/ui";

type MobileShellProps = {
  workspaceId: string;
};

const THREAD_ROUTE_ID = "/app/$workspaceId/$channelId/thread/$messageId";

const tabRoutes: Partial<Record<string, MobileTab>> = {
  "/app/$workspaceId/": "home",
  "/app/$workspaceId/activity": "activity",
  "/app/$workspaceId/dms": "dms",
  "/app/$workspaceId/me": "me",
};

/** モバイルの画面構成。下にボトムタブの画面を置き、チャンネルなどは右から重ねる。 右パネルの内容（スレッド・プロフィールなど）はさらにその上に重ねる */
export const MobileShell = ({ workspaceId }: MobileShellProps) => {
  const router = useRouter();
  const navigate = useNavigate();
  const canGoBack = useCanGoBack();
  const [lastTab, setLastTab] = useAtom(mobileTabAtom);
  const leafRouteId = useMatches({ select: (matches) => matches.at(-1)?.routeId ?? "" });
  // スレッドは右パネルとして重ねるため、その下の画面はチャンネルのまま動かさない
  const stackKey = useMatches({
    select: (matches) => matches.findLast((match) => match.routeId !== THREAD_ROUTE_ID)?.pathname,
  });
  const routeTab = tabRoutes[leafRouteId];
  const tab = routeTab ?? lastTab;
  const { close, content } = useRightPanel(workspaceId);
  const viewport = useVisualViewport();
  const showTabBar = routeTab !== undefined && !content && !viewport?.keyboardOpen;

  useEffect(() => {
    if (routeTab !== undefined) {
      setLastTab(routeTab);
    }
  }, [routeTab, setLastTab]);

  const back = () => {
    if (canGoBack) {
      router.history.back();
      return;
    }
    void navigate({ params: { workspaceId }, to: "/app/$workspaceId" });
  };

  return (
    <div
      data-keyboard={viewport?.keyboardOpen || undefined}
      style={{ height: viewport?.height }}
      className="group/shell flex h-full flex-col bg-surface pt-[env(safe-area-inset-top)] pr-[env(safe-area-inset-right)] pl-[env(safe-area-inset-left)] font-sans text-text [word-break:auto-phrase]"
    >
      <div className="relative flex min-h-0 flex-1 flex-col overflow-clip">
        {tab === "home" && <MobileHome workspaceId={workspaceId} />}
        {tab === "dms" && <DMsPage />}
        {tab === "activity" && <ActivityPage />}
        {tab === "me" && <MePage />}
        {routeTab === undefined && (
          <MobileStackLayer key={stackKey} onBack={back}>
            <Outlet />
          </MobileStackLayer>
        )}
        {content && (
          // パネルも URL で開くので、戻るは履歴を戻る。ディープリンクで開いたときだけ閉じる
          <MobileStackLayer key={content.key} onBack={canGoBack ? back : close}>
            <header className="flex h-12 shrink-0 items-center gap-1 border-b border-border px-3">
              <BackButton />
              <h2 className="m-0 min-w-0 flex-1 truncate text-[16px] font-bold">{content.title}</h2>
              {content.extra}
            </header>
            <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">{content.body}</div>
          </MobileStackLayer>
        )}
      </div>
      {/* タブの画面ではボトムタブの上に出す。チャンネルの画面では入力欄の上（ChannelPage） */}
      {showTabBar && <MiniPlayer variant="mobile" />}
      {showTabBar && <MobileTabBar workspaceId={workspaceId} />}
    </div>
  );
};
