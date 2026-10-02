import { useEffect } from "react";

import { useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { AppShell } from "#/features/layout/components/AppShell";
import { lastWorkspaceIdAtom } from "#/providers/store/workspace";
import { WsProvider } from "#/providers/ws/WsProvider";

export const WorkspaceLayout = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const setLastWorkspaceId = useSetAtom(lastWorkspaceIdAtom);

  useEffect(() => {
    setLastWorkspaceId(workspaceId);
  }, [workspaceId, setLastWorkspaceId]);

  return (
    <WsProvider workspaceId={workspaceId}>
      <AppShell workspaceId={workspaceId} />
    </WsProvider>
  );
};
