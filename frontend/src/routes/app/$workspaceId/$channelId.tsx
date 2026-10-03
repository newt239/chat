import { createFileRoute } from "@tanstack/react-router";

import { ChannelPage } from "#/features/channel/components/ChannelPage";
import { channelSearchSchema } from "#/features/channel/schemas";

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: ChannelPage,
  validateSearch: channelSearchSchema,
});
