import { useEffect } from "react";

import { useSetAtom } from "jotai";
import { Outlet } from "react-router";

import { useWorkspaceId } from "#/lib/routeParams";
import { syncCurrentWorkspaceAtom } from "#/providers/store/workspace";

export const WorkspaceLayout = () => {
  const workspaceId = useWorkspaceId();
  const syncCurrentWorkspace = useSetAtom(syncCurrentWorkspaceAtom);

  // これがないと /app/:workspaceId/:channelId への直接アクセスがワークスペース未選択扱いになる
  useEffect(() => {
    syncCurrentWorkspace(workspaceId);
  }, [workspaceId, syncCurrentWorkspace]);

  return <Outlet />;
};
