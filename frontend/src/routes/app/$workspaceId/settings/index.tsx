import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/app/$workspaceId/settings/")({
  beforeLoad: ({ params }) => {
    throw redirect({
      params: { ...params, section: "account" },
      replace: true,
      to: "/app/$workspaceId/settings/$section",
    });
  },
});
