import { IconSettings, IconUsers } from "@tabler/icons-react";
import { useSetAtom } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles";
import { NavLink } from "#/features/layout/components/NavLink";
import { navItemClassName } from "#/features/layout/utils/navTone";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

import { useUserGroups } from "../hooks/useUserGroups";

type UserGroupNavListProps = {
  workspaceId: string;
};

// サイドバーのユーザーグループ。押すと右パネルに詳細を開く
export const UserGroupNavList = ({ workspaceId }: UserGroupNavListProps) => {
  const { t } = useTranslation();
  const { data: groups = [] } = useUserGroups(workspaceId);
  const setRightPanel = useSetAtom(setRightSidePanelViewAtom);

  return (
    <>
      {groups.map((group) => (
        <Button
          key={group.id}
          onPress={() => {
            setRightPanel({ groupId: group.id, type: "user-group" });
          }}
          className={`${navItemClassName} ${focusRing}`}
        >
          <IconUsers aria-hidden />
          <span className="min-w-0 flex-1 truncate">@{group.name}</span>
        </Button>
      ))}
      <NavLink to="/app/$workspaceId/groups" params={{ workspaceId }}>
        <IconSettings aria-hidden />
        {t("userGroup.manage")}
      </NavLink>
    </>
  );
};
