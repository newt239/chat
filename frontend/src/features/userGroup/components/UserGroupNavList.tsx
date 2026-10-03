import { IconUsers } from "@tabler/icons-react";

import { NavLink } from "#/components/block/NavLink/NavLink";
import { openPanel } from "#/lib/overlaySearch";

import { useUserGroups } from "../hooks/useUserGroups";

type UserGroupNavListProps = {
  workspaceId: string;
};

// サイドバーのユーザーグループ。押すと右パネルに詳細を開く
export const UserGroupNavList = ({ workspaceId }: UserGroupNavListProps) => {
  const { data: groups = [] } = useUserGroups(workspaceId);

  return groups.map((group) => (
    <NavLink key={group.id} to="." search={openPanel({ group: group.id })}>
      <IconUsers aria-hidden />
      <span className="min-w-0 flex-1 truncate">@{group.name}</span>
    </NavLink>
  ));
};
