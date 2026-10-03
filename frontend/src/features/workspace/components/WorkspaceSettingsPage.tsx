import { IconAdjustments, IconBuilding, IconMoodSmile } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { SettingsLayout } from "#/components/block/SettingsLayout/SettingsLayout";
import { settingsNavLinkClassName } from "#/components/block/SettingsLayout/settingsNavLinkClassName";
import { Link } from "#/components/ui/Link/Link";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { CustomEmojiSettings } from "#/features/customEmoji/components/CustomEmojiSettings";

import { useWorkspaces } from "../hooks/useWorkspace";
import { findWorkspaceSettingsSection, workspaceSettingsSections } from "../schemas";
import { WorkspaceGeneralSettings } from "./WorkspaceGeneralSettings";

import type { WorkspaceSettingsSection } from "../schemas";

import type { Workspace } from "#/gen/chat/v1/workspace_service_pb";

const sectionIcons: Record<WorkspaceSettingsSection, typeof IconBuilding> = {
  emoji: IconMoodSmile,
  general: IconAdjustments,
};

const sectionBodies: Record<WorkspaceSettingsSection, (workspace: Workspace) => React.JSX.Element> =
  {
    emoji: (workspace) => <CustomEmojiSettings workspaceId={workspace.id} />,
    // 保存後に一覧が更新されてもフォームを作り直さない
    general: (workspace) => <WorkspaceGeneralSettings key={workspace.id} workspace={workspace} />,
  };

const workspaceSettingsRoute = getRouteApi("/app/$workspaceId/workspace-settings/{-$section}");

export const WorkspaceSettingsPage = () => {
  const { t } = useTranslation();
  const { section, workspaceId } = workspaceSettingsRoute.useParams();
  const current = findWorkspaceSettingsSection(section) ?? "general";
  const { data: workspaces } = useWorkspaces();
  const workspace = workspaces?.find((candidate) => candidate.id === workspaceId);

  return (
    <SettingsLayout
      icon={<IconBuilding />}
      title={t("workspace.settings.title")}
      sectionTitle={t(`workspace.settings.sections.${current}`)}
      nav={workspaceSettingsSections.map((name) => {
        const Icon = sectionIcons[name];
        return (
          <Link
            className={settingsNavLinkClassName}
            key={name}
            to="/app/$workspaceId/workspace-settings/{-$section}"
            params={{ section: name, workspaceId }}
            replace
          >
            <Icon aria-hidden />
            {t(`workspace.settings.sections.${name}`)}
          </Link>
        );
      })}
    >
      {workspace === undefined ? (
        <Skeleton className="h-64 w-full rounded-xl" />
      ) : (
        sectionBodies[current](workspace)
      )}
    </SettingsLayout>
  );
};
