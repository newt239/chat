import { IconAdjustments, IconBuilding, IconUsers } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { isAdminRole } from "#/features/admin/utils/isAdminRole";
import { SettingsLayout } from "#/features/settings/components/SettingsLayout";
import { SettingsNavLink } from "#/features/settings/components/SettingsNavLink";

import { useWorkspaces } from "../hooks/useWorkspace";
import { isWorkspaceSettingsSection, workspaceSettingsSections } from "../schemas";
import { WorkspaceGeneralSettings } from "./WorkspaceGeneralSettings";
import { WorkspaceMemberManager } from "./WorkspaceMemberManager";

import type { WorkspaceSettingsSection } from "../schemas";

const sectionIcons: Record<WorkspaceSettingsSection, typeof IconUsers> = {
  general: IconAdjustments,
  members: IconUsers,
};

const workspaceSettingsRoute = getRouteApi("/app/$workspaceId/workspace-settings/$section");

export const WorkspaceSettingsPage = () => {
  const { t } = useTranslation();
  const { section, workspaceId } = workspaceSettingsRoute.useParams();
  const current = isWorkspaceSettingsSection(section) ? section : "general";
  const { data: workspaces } = useWorkspaces();
  const workspace = workspaces?.find((candidate) => candidate.id === workspaceId);

  const renderBody = () => {
    if (workspace === undefined) {
      return <Skeleton className="h-64 w-full rounded-xl" />;
    }
    switch (current) {
      case "general": {
        // 保存後に一覧が更新されてもフォームを作り直さない
        return <WorkspaceGeneralSettings key={workspace.id} workspace={workspace} />;
      }
      case "members": {
        return (
          <WorkspaceMemberManager
            workspaceId={workspaceId}
            canManage={isAdminRole(workspace.role)}
          />
        );
      }
      default: {
        return null;
      }
    }
  };

  return (
    <SettingsLayout
      icon={<IconBuilding />}
      title={t("workspace.settings.title")}
      sectionTitle={t(`workspace.settings.sections.${current}`)}
      nav={workspaceSettingsSections.map((name) => {
        const Icon = sectionIcons[name];
        return (
          <SettingsNavLink
            key={name}
            to="/app/$workspaceId/workspace-settings/$section"
            params={{ section: name, workspaceId }}
            replace
          >
            <Icon aria-hidden />
            {t(`workspace.settings.sections.${name}`)}
          </SettingsNavLink>
        );
      })}
    >
      {renderBody()}
    </SettingsLayout>
  );
};
