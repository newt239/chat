import { IconLogout, IconSettings, IconUser } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { MenuSection } from "#/components/ui/MenuSection/MenuSection";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { focusRing } from "#/components/ui/styles/styles";
import { useLogout } from "#/features/auth/hooks/useLogout";
import { useMe } from "#/hooks/useMe";
import { openPanel } from "#/lib/overlaySearch";

// サイドバー下部の自分の名前。プロフィール・設定・ログアウトを出す
export const SidebarFooter = () => {
  const { t } = useTranslation();
  const { data: user } = useMe();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const logout = useLogout();

  if (user === undefined) {
    return null;
  }

  return (
    <div className="flex shrink-0 items-center gap-1 border-t border-(--nav-hover) px-2 py-1.5">
      <Menu
        placement="top start"
        trigger={
          <Button
            className={`flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md px-1.5 py-1 text-left data-hovered:bg-(--nav-hover) ${focusRing}`}
          >
            <Avatar name={user.displayName} src={user.avatarUrl} size={28} isOnline />
            <span className="min-w-0 truncate text-body-sm font-bold text-(--nav-strong)">
              {user.displayName}
            </span>
          </Button>
        }
      >
        <MenuSection title={user.email}>
          <MenuItemLink icon={<IconUser />} to="." search={openPanel({ profile: user.id })}>
            {t("shell.me.profile")}
          </MenuItemLink>
          <MenuItemLink
            icon={<IconSettings />}
            to="/app/$workspaceId/settings/{-$section}"
            params={{ section: "theme", workspaceId }}
          >
            {t("shell.me.settings")}
          </MenuItemLink>
        </MenuSection>
        <MenuSeparator />
        <MenuItem
          icon={<IconLogout />}
          onAction={() => {
            logout.mutate({});
          }}
        >
          {t("shell.me.logout")}
        </MenuItem>
      </Menu>
    </div>
  );
};
