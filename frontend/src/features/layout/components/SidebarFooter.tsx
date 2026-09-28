import { IconLogout, IconSettings, IconUser } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { IconButton } from "#/components/ui/IconButton";
import { Menu } from "#/components/ui/Menu";
import { MenuItem } from "#/components/ui/MenuItem";
import { MenuItemLink } from "#/components/ui/MenuItemLink";
import { MenuSection } from "#/components/ui/MenuSection";
import { MenuSeparator } from "#/components/ui/MenuSeparator";
import { focusRing } from "#/components/ui/styles";
import { useLogout } from "#/features/auth/hooks/useLogout";
import { userAtom } from "#/providers/store/auth";

import { openDialog, openPanel } from "../utils/overlaySearch";

// サイドバー下部の自分の名前。プロフィール・設定・ログアウトを出す
export const SidebarFooter = () => {
  const { t } = useTranslation();
  const user = useAtomValue(userAtom);
  const navigate = useNavigate();
  const logout = useLogout();

  if (user === null) {
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
            <Avatar name={user.displayName} src={user.avatarUrl} size={28} presence="online" />
            <span className="min-w-0 truncate text-[13px] font-bold text-(--nav-strong)">
              {user.displayName}
            </span>
          </Button>
        }
      >
        <MenuSection title={user.email}>
          <MenuItemLink icon={<IconUser />} to="." search={openPanel({ profile: user.id })}>
            {t("shell.me.profile")}
          </MenuItemLink>
          <MenuItemLink icon={<IconSettings />} to="." search={openDialog({ settings: "theme" })}>
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
      <IconButton
        label={t("shell.me.settings")}
        className="text-(--nav-muted) data-hovered:bg-(--nav-hover) data-hovered:text-(--nav-strong)"
        onPress={() => {
          void navigate({ search: openDialog({ settings: "theme" }), to: "." });
        }}
      >
        <IconSettings />
      </IconButton>
    </div>
  );
};
