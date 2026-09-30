import { createFileRoute, redirect } from "@tanstack/react-router";

import { SettingsPage } from "#/features/settings/components/SettingsPage";
import { isSettingsSection } from "#/features/settings/schemas";

export const Route = createFileRoute("/app/$workspaceId/settings/$section")({
  beforeLoad: ({ params }) => {
    if (!isSettingsSection(params.section)) {
      throw redirect({
        params: { ...params, section: "account" },
        replace: true,
        to: "/app/$workspaceId/settings/$section",
      });
    }
  },
  component: SettingsPage,
});
