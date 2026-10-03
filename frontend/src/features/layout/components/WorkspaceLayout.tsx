import { useParams } from "@tanstack/react-router";

import { AppShell } from "#/features/layout/components/AppShell";
import { WsProvider } from "#/providers/ws/WsProvider";

export const WorkspaceLayout = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });

  return (
    <WsProvider workspaceId={workspaceId}>
      <AppShell workspaceId={workspaceId} />
    </WsProvider>
  );
};
