import { useEffect } from "react";

import { createFileRoute, redirect } from "@tanstack/react-router";

import { ChatLayout } from "@/components/ChatLayout";
import { useAuthStore } from "@/lib/store/auth";
import { useWorkspaceStore } from "@/lib/store/workspace";

const ChannelComponent = () => {
  const { workspaceId, channelId } = Route.useParams();
  const setCurrentWorkspace = useWorkspaceStore((state) => state.setCurrentWorkspace);

  useEffect(() => {
    setCurrentWorkspace(workspaceId);
  }, [workspaceId, setCurrentWorkspace]);

  return <ChatLayout workspaceId={workspaceId} channelId={channelId} />;
};

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  beforeLoad: () => {
    const isAuthenticated = useAuthStore.getState().isAuthenticated;
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: ChannelComponent,
});
