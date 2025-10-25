import { useEffect, useRef } from "react";

import { useSetAtom } from "jotai";

import { WorkspaceList } from "@/features/workspace/components/WorkspaceList";
import { useWorkspaces } from "@/features/workspace/hooks/useWorkspace";
import { navigateToWorkspace } from "@/lib/navigation";
import { setCurrentWorkspaceAtom } from "@/lib/store/workspace";

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

    const workspaceId =
      parsed.state?.currentWorkspaceId ?? parsed.currentWorkspaceId;

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
        navigateToWorkspace(storedWorkspaceId);
        return;
      }
    }

    const firstWorkspace = workspaces[0];

    if (firstWorkspace) {
      hasRedirected.current = true;
      setCurrentWorkspace(firstWorkspace.id);
      navigateToWorkspace(firstWorkspace.id);
    }
  }, [setCurrentWorkspace, workspaces]);

  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">ワークスペースを選択してください</h1>
        <p className="text-gray-600 mb-8">
          参加しているワークスペースから選択するか、新しいワークスペースを作成してください。
        </p>
        <WorkspaceList />
      </div>
    </div>
  );
};
