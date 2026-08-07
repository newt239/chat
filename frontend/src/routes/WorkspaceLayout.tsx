import { useEffect } from "react";

import { useSetAtom } from "jotai";
import { Outlet } from "react-router";

import { useWorkspaceId } from "#/lib/routeParams";
import { syncCurrentWorkspaceAtom } from "#/providers/store/workspace";

export const WorkspaceLayout = () => {
  const workspaceId = useWorkspaceId();
  const syncCurrentWorkspace = useSetAtom(syncCurrentWorkspaceAtom);

  // URL を単一の情報源としてストアを追従させる。
  // これがないと /app/:workspaceId/:channelId へ直接アクセスした際にワークスペース未選択扱いになる。
  useEffect(() => {
    syncCurrentWorkspace(workspaceId);
  }, [workspaceId, syncCurrentWorkspace]);

  return <Outlet />;
};
