import { IconSettings, IconUsers } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { NavLink } from "#/features/layout/components/NavLink";
import { openPanel } from "#/features/layout/utils/overlaySearch";

import { useUserGroups } from "../hooks/useUserGroups";

type UserGroupNavListProps = {
  workspaceId: string;
};

// サイドバーのユーザーグループ。押すと右パネルに詳細を開く
export const UserGroupNavList = ({ workspaceId }: UserGroupNavListProps) => {
  const { t } = useTranslation();
  const { data: groups = [] } = useUserGroups(workspaceId);

  return (
    <>
      {groups.map((group) => (
        <NavLink key={group.id} to="." search={openPanel({ group: group.id })}>
          <IconUsers aria-hidden />
          <span className="min-w-0 flex-1 truncate">@{group.name}</span>
        </NavLink>
      ))}
      <NavLink to="/app/$workspaceId/groups" params={{ workspaceId }}>
        <IconSettings aria-hidden />
        {t("userGroup.manage")}
      </NavLink>
    </>
  );
};
