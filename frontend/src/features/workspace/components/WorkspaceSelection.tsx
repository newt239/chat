import { useEffect, useRef } from "react";

import { useAtomValue, useSetAtom } from "jotai";
import { useNavigate } from "react-router";

import { WorkspaceList } from "#/features/workspace/components/WorkspaceList";
import { useWorkspaces } from "#/features/workspace/hooks/useWorkspace";
import { paths } from "#/lib/paths";
import { currentWorkspaceIdAtom, setCurrentWorkspaceAtom } from "#/providers/store/workspace";

export const WorkspaceSelection = () => {
  const { data: workspaces } = useWorkspaces();
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);
  const storedWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const navigate = useNavigate();
  const hasRedirected = useRef(false);

  useEffect(() => {
    if (hasRedirected.current) {
      return;
    }

    if (!Array.isArray(workspaces) || workspaces.length === 0) {
      return;
    }

    if (storedWorkspaceId) {
      const storedExists = workspaces.some((workspace) => workspace.id === storedWorkspaceId);

      if (storedExists) {
        hasRedirected.current = true;
        setCurrentWorkspace(storedWorkspaceId);
        void navigate(paths.workspace(storedWorkspaceId));
        return;
      }
    }

    const [firstWorkspace] = workspaces;

    if (firstWorkspace) {
      hasRedirected.current = true;
      setCurrentWorkspace(firstWorkspace.id);
      void navigate(paths.workspace(firstWorkspace.id));
    }
  }, [setCurrentWorkspace, storedWorkspaceId, workspaces, navigate]);

  return (
    <div className="flex h-full items-center justify-center">
      <WorkspaceList />
    </div>
  );
};
