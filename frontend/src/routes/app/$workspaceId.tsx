import { useEffect } from "react";

import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { store } from "@/lib/store";
import { isAuthenticatedAtom } from "@/lib/store/auth";
import { setCurrentWorkspaceAtom } from "@/lib/store/workspace";

const WorkspaceComponent = () => {
  const { workspaceId } = Route.useParams();
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);

  useEffect(() => {
    setCurrentWorkspace(workspaceId);
  }, [workspaceId, setCurrentWorkspace]);

  return <Outlet />;
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
