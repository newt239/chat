import { createFileRoute } from "@tanstack/react-router";

import { ThreadListPage } from "@/features/thread/components/ThreadListPage";

export const Route = createFileRoute("/app/$workspaceId/threads")({
  component: ThreadListPage,
});
