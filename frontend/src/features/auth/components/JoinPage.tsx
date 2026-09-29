import { getRouteApi } from "@tanstack/react-router";

import { JoinWorkspace } from "#/features/auth/components/JoinWorkspace";

const joinRoute = getRouteApi("/join/$workspaceId");

export const JoinPage = () => {
  const { workspaceId } = joinRoute.useParams();
  return <JoinWorkspace workspaceId={workspaceId} />;
};
