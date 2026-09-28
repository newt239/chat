import { useEffect } from "react";

import { Outlet, useCanGoBack, useMatches, useNavigate, useRouter } from "@tanstack/react-router";
import { useAtom } from "jotai";

import { MiniPlayer } from "#/features/player/components/MiniPlayer";
import { ActivityPage } from "#/pages/ActivityPage";
import { DMsPage } from "#/pages/DMsPage";
import { MePage } from "#/pages/MePage";
import { mobileTabAtom } from "#/providers/store/ui";

import { useRightPanel } from "../hooks/useRightPanel";
import { BackButton } from "./BackButton";
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
    <div className="flex h-full flex-col bg-surface pt-[env(safe-area-inset-top)] font-sans text-text [word-break:auto-phrase]">
      <div className="relative flex min-h-0 flex-1 flex-col overflow-hidden">
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
          <MobileStackLayer
            key={content.key}
            onBack={leafRouteId === THREAD_ROUTE_ID && canGoBack ? back : close}
          >
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
      {routeTab !== undefined && !content && <MiniPlayer variant="mobile" />}
      {routeTab !== undefined && !content && <MobileTabBar workspaceId={workspaceId} />}
    </div>
  );
};
