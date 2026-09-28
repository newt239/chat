import { useEffect } from "react";

import { Outlet, useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { syncCurrentWorkspaceAtom } from "#/providers/store/workspace";

export const WorkspaceLayout = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const syncCurrentWorkspace = useSetAtom(syncCurrentWorkspaceAtom);

  // これがないと /app/:workspaceId/:channelId への直接アクセスがワークスペース未選択扱いになる
  useEffect(() => {
    syncCurrentWorkspace(workspaceId);
  }, [workspaceId, syncCurrentWorkspace]);

  return <Outlet />;
};
