import {
  IconChartBar,
  IconChevronDown,
  IconPlus,
  IconSettings,
  IconShieldCheck,
} from "@tabler/icons-react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { MenuSection } from "#/components/ui/MenuSection/MenuSection";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { focusRing } from "#/components/ui/styles/styles";
import { openDialog } from "#/features/layout/utils/overlaySearch";
import { useMyWorkspaceRole } from "#/hooks/useMyWorkspaceRole";
import { isAdminRole } from "#/lib/isAdminRole";

import { useWorkspaces } from "../hooks/useWorkspace";
import { WorkspaceLogo } from "./WorkspaceLogo";

type WorkspaceMenuProps = {
  workspaceId: string;
};

// サイドバー上部のワークスペース名。切り替え・設定・インサイトと管理画面への移動をまとめる
export const WorkspaceMenu = ({ workspaceId }: WorkspaceMenuProps) => {
  const { t } = useTranslation();
  const { data: workspaces = [] } = useWorkspaces();
  const isAdmin = isAdminRole(useMyWorkspaceRole(workspaceId).data);
  const current = workspaces.find((workspace) => workspace.id === workspaceId);
  const name = current?.name ?? workspaceId;

  return (
    <Menu
      placement="bottom start"
      trigger={
        <Button
          className={`flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md px-1.5 py-1 text-left text-(--nav-strong) data-hovered:bg-(--nav-hover) ${focusRing}`}
        >
          <WorkspaceLogo name={name} iconUrl={current?.iconUrl} />
          <span className="min-w-0 truncate text-[15px] font-bold">{name}</span>
          <IconChevronDown aria-hidden className="size-3.5 shrink-0 text-(--nav-muted)" />
        </Button>
      }
    >
      <MenuSection title={t("shell.workspace.switch")}>
        {workspaces.map((workspace) => (
          <MenuItemLink
            key={workspace.id}
            to="/app/$workspaceId"
            params={{ workspaceId: workspace.id }}
            icon={<WorkspaceLogo name={workspace.name} iconUrl={workspace.iconUrl} />}
            shortcut={workspace.id === workspaceId ? "✓" : undefined}
          >
            {workspace.name}
          </MenuItemLink>
        ))}
        <MenuItemLink
          icon={<IconPlus />}
          to="."
          search={openDialog({ dialog: "create-workspace" })}
        >
          {t("shell.workspace.create")}
        </MenuItemLink>
      </MenuSection>
      <MenuSeparator />
      <MenuItemLink
        to="/app/$workspaceId/insights"
        params={{ workspaceId }}
        icon={<IconChartBar />}
      >
        {t("shell.nav.insights")}
      </MenuItemLink>
      {isAdmin && (
        <MenuItemLink
          to="/app/$workspaceId/admin"
          params={{ workspaceId }}
          icon={<IconShieldCheck />}
        >
          {t("shell.nav.admin")}
        </MenuItemLink>
      )}
      {current && (
        <MenuItemLink
          icon={<IconSettings />}
          to="/app/$workspaceId/workspace-settings/$section"
          params={{ section: "general", workspaceId }}
        >
          {t("shell.workspace.settings")}
        </MenuItemLink>
      )}
    </Menu>
  );
};
