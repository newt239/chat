import { useEffect, useRef } from "react";

import { useSetAtom } from "jotai";

import { WorkspaceList } from "@/features/workspace/components/WorkspaceList";
import { useWorkspaces } from "@/features/workspace/hooks/useWorkspace";
import { router } from "@/lib/router";
import { setCurrentWorkspaceAtom } from "@/providers/store/workspace";

type WorkspaceStorageState = {
  state?: {
    currentWorkspaceId?: string | null;
  };
};

const getStoredWorkspaceId = () => {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const stored = localStorage.getItem("workspace-storage");
    if (!stored) {
      return null;
    }

    const parsed = JSON.parse(stored) as WorkspaceStorageState & {
      currentWorkspaceId?: string | null;
    };

    const workspaceId = parsed.state?.currentWorkspaceId ?? parsed.currentWorkspaceId;

    if (typeof workspaceId === "string" && workspaceId.length > 0) {
      return workspaceId;
    }
  } catch (error) {
    console.warn("ワークスペース情報の取得に失敗しました", error);
  }

  return null;
};

export const WorkspaceSelection = () => {
  const { data: workspaces } = useWorkspaces();
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);
  const hasRedirected = useRef(false);

  useEffect(() => {
    if (hasRedirected.current) {
      return;
    }

    if (!Array.isArray(workspaces) || workspaces.length === 0) {
      return;
    }

    const storedWorkspaceId = getStoredWorkspaceId();

    if (storedWorkspaceId) {
      const storedExists = workspaces.some((workspace) => workspace.id === storedWorkspaceId);

      if (storedExists) {
        hasRedirected.current = true;
        setCurrentWorkspace(storedWorkspaceId);
        router.navigate({ to: "/app/$workspaceId", params: { workspaceId: storedWorkspaceId } });
        return;
      }
    }

    const firstWorkspace = workspaces[0];

    if (firstWorkspace) {
      hasRedirected.current = true;
      setCurrentWorkspace(firstWorkspace.id);
      router.navigate({ to: "/app/$workspaceId", params: { workspaceId: firstWorkspace.id } });
    }
  }, [setCurrentWorkspace, workspaces]);

  return (
    <div className="flex h-full items-center justify-center">
      <WorkspaceList />
    </div>
  );
};
