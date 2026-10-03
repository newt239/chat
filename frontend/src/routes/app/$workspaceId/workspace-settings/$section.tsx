import { createFileRoute, redirect } from "@tanstack/react-router";

import { WorkspaceSettingsPage } from "#/features/workspace/components/WorkspaceSettingsPage";
import { findWorkspaceSettingsSection } from "#/features/workspace/schemas";

export const Route = createFileRoute("/app/$workspaceId/workspace-settings/$section")({
  beforeLoad: ({ params }) => {
    if (findWorkspaceSettingsSection(params.section) === undefined) {
      throw redirect({
        params: { ...params, section: "general" },
        replace: true,
        to: "/app/$workspaceId/workspace-settings/$section",
      });
    }
  },
  component: WorkspaceSettingsPage,
});
