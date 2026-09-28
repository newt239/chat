import { IconLogout, IconSettings, IconUser } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { IconButton } from "#/components/ui/IconButton";
import { Menu } from "#/components/ui/Menu";
import { MenuItem } from "#/components/ui/MenuItem";
import { MenuSection } from "#/components/ui/MenuSection";
import { MenuSeparator } from "#/components/ui/MenuSeparator";
import { focusRing } from "#/components/ui/styles";
import { useLogout } from "#/features/auth/hooks/useLogout";
import { userAtom } from "#/providers/store/auth";
import { setRightSidePanelViewAtom, settingsSectionAtom } from "#/providers/store/ui";

// サイドバー下部の自分の名前。プロフィール・設定・ログアウトを出す
export const SidebarFooter = () => {
  const { t } = useTranslation();
  const user = useAtomValue(userAtom);
  const setRightPanel = useSetAtom(setRightSidePanelViewAtom);
  const logout = useLogout();
  const setSettingsSection = useSetAtom(settingsSectionAtom);

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
          <MenuItem
            icon={<IconUser />}
            onAction={() => {
              setRightPanel({ type: "user-profile", userId: user.id });
            }}
          >
            {t("shell.me.profile")}
          </MenuItem>
          <MenuItem
            icon={<IconSettings />}
            onAction={() => {
              setSettingsSection("theme");
            }}
          >
            {t("shell.me.settings")}
          </MenuItem>
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
          setSettingsSection("theme");
        }}
      >
        <IconSettings />
      </IconButton>
    </div>
  );
};
