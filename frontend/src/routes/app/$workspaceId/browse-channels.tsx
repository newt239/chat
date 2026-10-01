import { createFileRoute } from "@tanstack/react-router";

import { BrowseChannelsPage } from "#/features/channel/components/BrowseChannelsPage";
import { browseChannelsSearchSchema } from "#/features/channel/schemas";

export const Route = createFileRoute("/app/$workspaceId/browse-channels")({
  component: BrowseChannelsPage,
  validateSearch: browseChannelsSearchSchema,
});
