import { useEffect } from "react";

import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { ChannelList } from "@/features/channel/components/ChannelList";
import { store } from "@/lib/store";
import { isAuthenticatedAtom } from "@/lib/store/auth";
import { setCurrentWorkspaceAtom } from "@/lib/store/workspace";

const WorkspaceComponent = () => {
  const { workspaceId } = Route.useParams();
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);

  useEffect(() => {
    setCurrentWorkspace(workspaceId);
  }, [workspaceId, setCurrentWorkspace]);

  return (
    <div className="grid h-full min-h-0 gap-6 lg:grid-cols-[320px_1fr]">
      <div className="space-y-6">
        <ChannelList workspaceId={workspaceId} />
      </div>
      <div className="min-h-0 w-full">
        <Outlet />
      </div>
    </div>
  );
};

export const Route = createFileRoute("/app/$workspaceId")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: WorkspaceComponent,
});
