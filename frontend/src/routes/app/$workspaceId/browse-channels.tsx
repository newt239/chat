import { createFileRoute } from "@tanstack/react-router";

import { BrowseChannelsPage } from "#/features/channel/components/BrowseChannelsPage";

export const Route = createFileRoute("/app/$workspaceId/browse-channels")({
  component: BrowseChannelsPage,
});
