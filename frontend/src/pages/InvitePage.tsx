import { getRouteApi } from "@tanstack/react-router";

import { InvitationAccept } from "#/features/auth/components/InvitationAccept";

const inviteRoute = getRouteApi("/invite/$token");

export const InvitePage = () => {
  const { token } = inviteRoute.useParams();
  return <InvitationAccept token={token} />;
};
