import { useEffect, useRef, useState } from "react";

import { Outlet, useLocation, useMatches, useNavigate, useRouter } from "@tanstack/react-router";

import { BackButton } from "#/components/block/BackButton/BackButton";
import { DMsPage } from "#/features/channel/components/DMsPage";
import { MentionsPage } from "#/features/inbox/components/MentionsPage";
import { useVisualViewport } from "#/features/layout/hooks/useVisualViewport";

import { useRightPanel } from "../hooks/useRightPanel";
import { mobileTabs } from "../utils/mobileTabs";
import { MePage } from "./MePage";
import { MobileHome } from "./MobileHome";
import { MobileStackLayer } from "./MobileStackLayer";
import { MobileTabBar } from "./MobileTabBar";

import type { MobileTab } from "../utils/mobileTabs";

type MobileShellProps = {
  workspaceId: string;
};

const THREAD_ROUTE_ID = "/app/$workspaceId/$channelId/thread/$messageId";

// 下にボトムタブの画面を置き、チャンネルなどは右から、右パネルの内容はさらにその上に重ねる
export const MobileShell = ({ workspaceId }: MobileShellProps) => {
  const router = useRouter();
  const navigate = useNavigate();
  // oxlint-disable-next-line no-underscore-dangle -- TanStack Router が履歴に持たせる位置
  const historyIndex = useLocation({ select: (location) => location.state.__TSR_index });
  // ホームはインデックスのルートなので、末尾の / を落としてタブのパスと比べる
  const leafPath = useMatches({
    select: (matches) => matches.at(-1)?.routeId.replace(/\/$/u, "") ?? "",
  });
  // スレッドは右パネルとして重ねるため、その下の画面はチャンネルのまま動かさない
  const stackKey = useMatches({
    select: (matches) => matches.findLast((match) => match.routeId !== THREAD_ROUTE_ID)?.pathname,
  });
  const routeTab = mobileTabs.find((candidate) => candidate.to === leafPath);
  // チャンネルなどを重ねている間は、最後に開いたタブを下に残す
  const [lastTab, setLastTab] = useState<MobileTab>(mobileTabs[0]);
  if (routeTab !== undefined && routeTab !== lastTab) {
    setLastTab(routeTab);
  }
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
    visitedRef.current.set(historyIndex, { isTab: routeTab !== undefined, panelKey, stackKey });
  }, [historyIndex, routeTab, panelKey, stackKey]);

  // 退いた画面の下に見えていた画面へ移る。履歴の一つ前がその画面なら履歴を戻り、違えば置き換える
  const backToTab = () => {
    if (visitedRef.current.get(historyIndex - 1)?.isTab) {
      router.history.back();
      return;
    }
    void navigate({ params: { workspaceId }, replace: true, to: tab.to });
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
            {tab.name === "home" && <MobileHome workspaceId={workspaceId} />}
            {tab.name === "dms" && <DMsPage />}
            {tab.name === "activity" && <MentionsPage />}
            {tab.name === "me" && <MePage />}
          </div>
          {!viewport?.keyboardOpen && <MobileTabBar workspaceId={workspaceId} tab={tab} />}
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
