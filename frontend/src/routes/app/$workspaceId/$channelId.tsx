import { createFileRoute } from "@tanstack/react-router";

import { MessagePanel } from "@/features/message/components/MessagePanel";

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: MessagePanel,
});
