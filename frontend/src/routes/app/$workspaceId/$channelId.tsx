import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { ChannelPage } from "#/pages/ChannelPage";

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: ChannelPage,
  validateSearch: z.object({ message: z.string().optional().catch(undefined) }),
});
