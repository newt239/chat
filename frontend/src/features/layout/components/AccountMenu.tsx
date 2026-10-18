import { useState } from "react";

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
import { LogoutConfirmDialog } from "#/features/auth/components/LogoutConfirmDialog";
import { useMe } from "#/hooks/useMe";
import { openPanel } from "#/lib/overlaySearch";

// トップバー右端の自分のアバター。プロフィール・設定・ログアウトを出す
export const AccountMenu = () => {
  const { t } = useTranslation();
  const { data: user } = useMe();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  if (user === undefined) {
    return null;
  }

  return (
    <>
      <Menu
        placement="bottom end"
        trigger={
          <Button
            aria-label={t("shell.me.menu")}
            className={`flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-full data-hovered:bg-hover ${focusRing}`}
          >
            <Avatar name={user.displayName} src={user.avatarUrl} size={30} isOnline />
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
            setIsLoggingOut(true);
          }}
        >
          {t("shell.me.logout")}
        </MenuItem>
      </Menu>
      <LogoutConfirmDialog isOpen={isLoggingOut} onOpenChange={setIsLoggingOut} />
    </>
  );
};
