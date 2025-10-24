import { useEffect } from "react";

import { Outlet } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import { WorkspaceRightSidebar } from "./WorkspaceRightSidebar";

import { ChannelList } from "@/features/channel/components/ChannelList";
import { rightSidebarViewAtom } from "@/lib/store/ui";
import { setCurrentWorkspaceAtom } from "@/lib/store/workspace";

type WorkspaceComponentProps = {
  workspaceId: string;
};

export const WorkspaceComponent = ({ workspaceId }: WorkspaceComponentProps) => {
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);
  const rightSidebarView = useAtomValue(rightSidebarViewAtom);

  useEffect(() => {
    setCurrentWorkspace(workspaceId);
  }, [workspaceId, setCurrentWorkspace]);

  const layoutClassName =
    rightSidebarView.type === "hidden"
      ? "grid h-full min-h-0 gap-6 lg:grid-cols-[320px_1fr]"
      : "grid h-full min-h-0 gap-6 lg:grid-cols-[320px_1fr_280px]";

  return (
    <div className={layoutClassName}>
      <div className="space-y-6">
        <ChannelList workspaceId={workspaceId} />
      </div>
      <div className="min-h-0 w-full">
        <Outlet />
      </div>
      {rightSidebarView.type !== "hidden" && (
        <WorkspaceRightSidebar workspaceId={workspaceId} view={rightSidebarView} />
      )}
    </div>
  );
};
