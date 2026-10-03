import { Outlet, useParams } from "@tanstack/react-router";

import { useChannelRealtimeSync } from "#/features/channel/hooks/useChannelRealtimeSync";
import { useChannelViewersSync } from "#/features/channel/hooks/useChannelViewers";
import { useCustomEmojiRealtimeSync } from "#/features/customEmoji/hooks/useCustomEmojiRealtimeSync";
import { useDesktopNotifications } from "#/features/notification/hooks/useDesktopNotifications";
import { useSyncPushToken } from "#/features/notification/hooks/usePushNotifications";
import { useIsMobile } from "#/hooks/useMediaQuery";

import { useAppBadge } from "../hooks/useAppBadge";
import { useGlobalShortcuts } from "../hooks/useGlobalShortcuts";
import { MobileShell } from "./MobileShell";
import { RightSidePanel } from "./RightSidePanel";
import { Sidebar } from "./Sidebar";
import { WorkspaceDialogs } from "./WorkspaceDialogs";

type AppShellProps = {
  workspaceId: string;
};

// ワークスペース内の画面の枠。デスクトップは左にサイドバー、中央に各ページ、右にパネルを置く
export const AppShell = ({ workspaceId }: AppShellProps) => {
  const isMobile = useIsMobile();
  const currentChannelId = useParams({
    select: (params) => params.channelId ?? null,
    strict: false,
  });
  useChannelRealtimeSync(workspaceId, currentChannelId);
  useDesktopNotifications(workspaceId, currentChannelId);
  useGlobalShortcuts(workspaceId);
  useChannelViewersSync();
  useCustomEmojiRealtimeSync(workspaceId);
  useAppBadge(workspaceId);
  useSyncPushToken();

  return (
    <>
      {isMobile ? (
        <MobileShell workspaceId={workspaceId} />
      ) : (
        <div className="flex h-full overflow-clip bg-bg pr-[env(safe-area-inset-right)] pl-[env(safe-area-inset-left)] font-sans text-text">
          <Sidebar workspaceId={workspaceId} />
          <main className="flex min-h-0 min-w-0 flex-1 flex-col bg-surface">
            <Outlet />
          </main>
          <RightSidePanel workspaceId={workspaceId} />
        </div>
      )}
      <WorkspaceDialogs workspaceId={workspaceId} />
    </>
  );
};
