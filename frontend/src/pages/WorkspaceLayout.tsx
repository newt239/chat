import { useEffect } from "react";

import { useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { AppShell } from "#/features/layout/components/AppShell";
import { syncCurrentWorkspaceAtom } from "#/providers/store/workspace";

export const WorkspaceLayout = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const syncCurrentWorkspace = useSetAtom(syncCurrentWorkspaceAtom);

  // これがないと /app/:workspaceId/:channelId への直接アクセスがワークスペース未選択扱いになる
  useEffect(() => {
    syncCurrentWorkspace(workspaceId);
  }, [workspaceId, syncCurrentWorkspace]);

  return <AppShell workspaceId={workspaceId} />;
};
