import { useEffect, useRef } from "react";

import { Outlet, useLocation, useMatches, useNavigate, useRouter } from "@tanstack/react-router";
import { useAtom } from "jotai";

import { BackButton } from "#/components/block/BackButton/BackButton";
import { DMsPage } from "#/features/dm/components/DMsPage";
import { useVisualViewport } from "#/features/layout/hooks/useVisualViewport";
import { MiniPlayer } from "#/features/player/components/MiniPlayer";
import { mobileTabAtom } from "#/providers/store/ui";

import { useRightPanel } from "../hooks/useRightPanel";
import { ActivityPage } from "./ActivityPage";
import { MePage } from "./MePage";
import { MobileHome } from "./MobileHome";
import { MobileStackLayer } from "./MobileStackLayer";
import { MobileTabBar } from "./MobileTabBar";

import type { MobileTab } from "#/providers/store/ui";

type MobileShellProps = {
  workspaceId: string;
};

const THREAD_ROUTE_ID = "/app/$workspaceId/$channelId/thread/$messageId";

const tabPaths = {
  activity: "/app/$workspaceId/activity",
  dms: "/app/$workspaceId/dms",
  home: "/app/$workspaceId",
  me: "/app/$workspaceId/me",
} as const satisfies Record<MobileTab, string>;

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
  // oxlint-disable-next-line no-underscore-dangle -- TanStack Router が履歴に持たせる位置
  const historyIndex = useLocation({ select: (location) => location.state.__TSR_index });
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
  const isTabScreen = routeTab !== undefined && !content;

  const panelKey = content?.key ?? null;
  // 履歴の位置ごとの画面。戻る先が下に見えている画面と同じかを確かめるのに使う
  const visitedRef = useRef(
    new Map<number, { isTab: boolean; stackKey: string | undefined; panelKey: string | null }>(),
  );

  useEffect(() => {
    if (routeTab !== undefined) {
      setLastTab(routeTab);
    }
  }, [routeTab, setLastTab]);

  useEffect(() => {
    visitedRef.current.set(historyIndex, { isTab: routeTab !== undefined, panelKey, stackKey });
  }, [historyIndex, routeTab, panelKey, stackKey]);

  // 退いた画面の下に見えていた画面へ移る。履歴の一つ前がその画面なら履歴を戻り、違えば置き換える
  const backToTab = () => {
    if (visitedRef.current.get(historyIndex - 1)?.isTab) {
      router.history.back();
      return;
    }
    void navigate({ params: { workspaceId }, replace: true, to: tabPaths[tab] });
  };
  const backFromPanel = () => {
    const previous = visitedRef.current.get(historyIndex - 1);
    if (previous !== undefined && previous.stackKey === stackKey && previous.panelKey === null) {
      router.history.back();
      return;
    }
    close();
  };

  return (
    <div
      data-keyboard={viewport?.keyboardOpen || undefined}
      style={{ height: viewport?.height }}
      className="group/shell flex h-full flex-col bg-surface pt-[env(safe-area-inset-top)] pr-[env(safe-area-inset-right)] pl-[env(safe-area-inset-left)] font-sans text-text [word-break:auto-phrase]"
    >
      {/* 重ねた画面はボトムタブごと覆う。ボトムタブは下に置いたままにし、戻ったときに後から出てこないようにする */}
      <div className="relative flex min-h-0 flex-1 flex-col overflow-clip">
        <div inert={!isTabScreen} className="flex min-h-0 flex-1 flex-col">
          <div className="flex min-h-0 flex-1 flex-col">
            {tab === "home" && <MobileHome workspaceId={workspaceId} />}
            {tab === "dms" && <DMsPage />}
            {tab === "activity" && <ActivityPage />}
            {tab === "me" && <MePage />}
          </div>
          {/* タブの画面ではボトムタブの上に出す。チャンネルの画面では入力欄の上（ChannelPage） */}
          {isTabScreen && !viewport?.keyboardOpen && <MiniPlayer variant="mobile" />}
          {!viewport?.keyboardOpen && <MobileTabBar workspaceId={workspaceId} />}
        </div>
        {routeTab === undefined && (
          <MobileStackLayer key={stackKey} onBack={backToTab}>
            <Outlet />
          </MobileStackLayer>
        )}
        {content && (
          <MobileStackLayer key={content.key} onBack={backFromPanel}>
            <header className="flex h-12 shrink-0 items-center gap-1 border-b border-border px-3">
              <BackButton />
              <h2 className="m-0 min-w-0 flex-1 truncate text-title font-bold">{content.title}</h2>
              {content.extra}
            </header>
            <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">{content.body}</div>
          </MobileStackLayer>
        )}
      </div>
    </div>
  );
};
