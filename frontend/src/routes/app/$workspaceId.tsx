import { useEffect } from "react";

import { createFileRoute, redirect } from "@tanstack/react-router";

import { ChatLayout } from "@/features/message/components/ChatLayout";
import { navigateToChannel } from "@/lib/navigation";
import { useAuthStore } from "@/lib/store/auth";
import { useWorkspaceStore } from "@/lib/store/workspace";

const WorkspaceComponent = () => {
  const { workspaceId } = Route.useParams();
  const setCurrentWorkspace = useWorkspaceStore((state) => state.setCurrentWorkspace);
  const currentChannelId = useWorkspaceStore((state) => state.currentChannelId);

  useEffect(() => {
    setCurrentWorkspace(workspaceId);
  }, [workspaceId, setCurrentWorkspace]);

  // チャンネルが選択されている場合はそのチャンネルにリダイレクト
  useEffect(() => {
    if (currentChannelId) {
      navigateToChannel(workspaceId, currentChannelId);
    }
  }, [workspaceId, currentChannelId]);

  return <ChatLayout workspaceId={workspaceId} channelId={currentChannelId} />;
};

export const Route = createFileRoute("/app/$workspaceId")({
  beforeLoad: () => {
    const isAuthenticated = useAuthStore.getState().isAuthenticated;
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: WorkspaceComponent,
});
