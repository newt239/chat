import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { ChannelPage } from "#/features/channel/components/ChannelPage";
import { jumpDateSchema } from "#/features/message/utils/dateJump";

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: ChannelPage,
  validateSearch: z.object({
    date: jumpDateSchema.optional().catch(undefined),
    message: z.string().optional().catch(undefined),
  }),
});
