import { getRouteApi, Navigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";

import { WorkspaceList } from "#/features/workspace/components/WorkspaceList";
import { useWorkspaces } from "#/features/workspace/hooks/useWorkspace";
import { lastWorkspaceIdAtom } from "#/providers/store/workspace";

const shortcutRoutes = {
  activity: "/app/$workspaceId/mentions",
  dms: "/app/$workspaceId/dms",
  search: "/app/$workspaceId/search",
} as const;

const appIndexRoute = getRouteApi("/app/");

// 前回開いていた（なければ最初の）ワークスペースへ移動する。参加していなければ一覧を出す
export const WorkspaceSelection = () => {
  const { data: workspaces } = useWorkspaces();
  const open = appIndexRoute.useSearch({ select: (search) => search.open });
  const storedWorkspaceId = useAtomValue(lastWorkspaceIdAtom);
  const target =
    workspaces?.find((workspace) => workspace.id === storedWorkspaceId) ?? workspaces?.[0];

  if (target) {
    return (
      <Navigate
        to={open ? shortcutRoutes[open] : "/app/$workspaceId"}
        params={{ workspaceId: target.id }}
        replace
      />
    );
  }

  return (
    <main className="flex min-h-full justify-center bg-bg px-4 py-10 font-sans text-text">
      {workspaces && <WorkspaceList />}
    </main>
  );
};
