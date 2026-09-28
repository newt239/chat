import { Outlet, useParams } from "@tanstack/react-router";

import { useChannelRealtimeSync } from "#/features/channel/hooks/useChannelRealtimeSync";
import { useChannelViewersSync } from "#/features/channel/hooks/useChannelViewers";
import { useDMRealtimeSync } from "#/features/dm/hooks/useDMRealtimeSync";
import { useNotificationSync } from "#/features/notification/hooks/useNotificationSync";
import { SettingsDialog } from "#/features/settings/components/SettingsDialog";
import { useDesktopNotifications } from "#/features/settings/hooks/useDesktopNotifications";
import { useIsMobile } from "#/lib/useMediaQuery";

import { useGlobalShortcuts } from "../hooks/useGlobalShortcuts";
import { MobileShell } from "./MobileShell";
import { RightSidePanel } from "./RightSidePanel";
import { Sidebar } from "./Sidebar";

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
  useDMRealtimeSync(workspaceId, currentChannelId);
  useNotificationSync(workspaceId, currentChannelId);
  useDesktopNotifications(workspaceId, currentChannelId);
  useGlobalShortcuts(workspaceId);
  useChannelViewersSync();

  return (
    <>
      {isMobile ? (
        <MobileShell workspaceId={workspaceId} />
      ) : (
        <div className="flex h-full bg-bg font-sans text-text">
          <Sidebar workspaceId={workspaceId} />
          <main className="flex min-w-0 flex-1 flex-col bg-surface">
            <Outlet />
          </main>
          <RightSidePanel workspaceId={workspaceId} />
        </div>
      )}
      <SettingsDialog />
    </>
  );
};
