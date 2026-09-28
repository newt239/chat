import { useState } from "react";

import {
  IconChartBar,
  IconChevronDown,
  IconPlus,
  IconSettings,
  IconShieldCheck,
} from "@tabler/icons-react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Menu } from "#/components/ui/Menu";
import { MenuItem } from "#/components/ui/MenuItem";
import { MenuItemLink } from "#/components/ui/MenuItemLink";
import { MenuSection } from "#/components/ui/MenuSection";
import { MenuSeparator } from "#/components/ui/MenuSeparator";
import { focusRing } from "#/components/ui/styles";

import { useIsWorkspaceAdmin } from "../hooks/useIsWorkspaceAdmin";
import { useWorkspaces } from "../hooks/useWorkspace";
import { CreateWorkspaceModal } from "./CreateWorkspaceModal";
import { WorkspaceLogo } from "./WorkspaceLogo";
import { WorkspaceSettingsModal } from "./WorkspaceSettingsModal";

type WorkspaceMenuProps = {
  workspaceId: string;
};

// サイドバー上部のワークスペース名。切り替え・設定・インサイトと管理画面への移動をまとめる
export const WorkspaceMenu = ({ workspaceId }: WorkspaceMenuProps) => {
  const { t } = useTranslation();
  const { data: workspaces = [] } = useWorkspaces();
  const isAdmin = useIsWorkspaceAdmin(workspaceId);
  const [dialog, setDialog] = useState<"settings" | "create" | null>(null);
  const current = workspaces.find((workspace) => workspace.id === workspaceId);
  const name = current?.name ?? workspaceId;

  return (
    <>
      <Menu
        placement="bottom start"
        trigger={
          <Button
            className={`flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md px-1.5 py-1 text-left text-(--nav-strong) data-hovered:bg-(--nav-hover) ${focusRing}`}
          >
            <WorkspaceLogo name={name} />
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
              icon={<WorkspaceLogo name={workspace.name} />}
              shortcut={workspace.id === workspaceId ? "✓" : undefined}
            >
              {workspace.name}
            </MenuItemLink>
          ))}
          <MenuItem
            icon={<IconPlus />}
            onAction={() => {
              setDialog("create");
            }}
          >
            {t("shell.workspace.create")}
          </MenuItem>
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
          <MenuItem
            icon={<IconSettings />}
            onAction={() => {
              setDialog("settings");
            }}
          >
            {t("shell.workspace.settings")}
          </MenuItem>
        )}
      </Menu>
      <CreateWorkspaceModal
        isOpen={dialog === "create"}
        onOpenChange={(isOpen) => {
          setDialog(isOpen ? "create" : null);
        }}
      />
      {current && (
        <WorkspaceSettingsModal
          isOpen={dialog === "settings"}
          onOpenChange={(isOpen) => {
            setDialog(isOpen ? "settings" : null);
          }}
          workspace={current}
        />
      )}
    </>
  );
};
