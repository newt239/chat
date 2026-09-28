import { Navigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";

import { WorkspaceList } from "#/features/workspace/components/WorkspaceList";
import { useWorkspaces } from "#/features/workspace/hooks/useWorkspace";
import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

// 前回開いていた（なければ最初の）ワークスペースへ移動する。参加していなければ一覧を出す
export const WorkspaceSelection = () => {
  const { data: workspaces } = useWorkspaces();
  const storedWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const target =
    workspaces?.find((workspace) => workspace.id === storedWorkspaceId) ?? workspaces?.[0];

  if (target) {
    return <Navigate to="/app/$workspaceId" params={{ workspaceId: target.id }} replace />;
  }

  return (
    <main className="flex min-h-full justify-center bg-bg px-4 py-10 font-sans text-text">
      {workspaces && <WorkspaceList />}
    </main>
  );
};
