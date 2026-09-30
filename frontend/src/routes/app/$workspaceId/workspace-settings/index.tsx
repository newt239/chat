import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/app/$workspaceId/workspace-settings/")({
  beforeLoad: ({ params }) => {
    throw redirect({
      params: { ...params, section: "general" },
      replace: true,
      to: "/app/$workspaceId/workspace-settings/$section",
    });
  },
});
