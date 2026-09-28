import {
  IconBell,
  IconBookmark,
  IconChartBar,
  IconChevronRight,
  IconKey,
  IconKeyboard,
  IconLanguage,
  IconLogout,
  IconMessages,
  IconPalette,
  IconShieldCheck,
  IconUser,
} from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { cn, focusRing } from "#/components/ui/styles";
import { useLogout } from "#/features/auth/hooks/useLogout";
import { NavLink } from "#/features/layout/components/NavLink";
import { PageHeader } from "#/features/layout/components/PageHeader";
import { mobileNavTone, navItemClassName } from "#/features/layout/utils/navTone";
import { openDialog, openPanel } from "#/features/layout/utils/overlaySearch";
import { useIsWorkspaceAdmin } from "#/features/workspace/hooks/useIsWorkspaceAdmin";
import { userAtom } from "#/providers/store/auth";

import type { SettingsSection } from "#/features/layout/schemas";

const settingRows: [SettingsSection, typeof IconKey][] = [
  ["account", IconKey],
  ["notifications", IconBell],
  ["theme", IconPalette],
  ["display", IconLanguage],
  ["shortcuts", IconKeyboard],
];

const rowClassName = cn(navItemClassName, focusRing);

// モバイルの「自分」タブ。プロフィール・よく使う一覧・設定の入口をまとめる
export const MePage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const user = useAtomValue(userAtom);
  const isAdmin = useIsWorkspaceAdmin(workspaceId);
  const logout = useLogout();
  const params = { workspaceId };

  return (
    <>
      <PageHeader icon={<IconUser />} title={t("shell.tabs.me")} />
      <div className={cn(mobileNavTone, "flex min-h-0 flex-1 flex-col overflow-y-auto p-1.5")}>
        {user && (
          <NavLink to="." search={openPanel({ profile: user.id })} className="h-auto gap-3 py-3">
            <Avatar name={user.displayName} src={user.avatarUrl} size={52} presence="online" />
            <span className="flex min-w-0 flex-1 flex-col">
              <b className="truncate text-[16px]">{user.displayName}</b>
              <span className="truncate text-caption text-muted">{user.email}</span>
            </span>
            <IconChevronRight aria-hidden />
          </NavLink>
        )}
        <NavLink to="/app/$workspaceId/threads" params={params}>
          <IconMessages aria-hidden />
          {t("shell.nav.threads")}
        </NavLink>
        <NavLink to="/app/$workspaceId/bookmarks" params={params}>
          <IconBookmark aria-hidden />
          {t("shell.nav.bookmarks")}
        </NavLink>
        <NavLink to="/app/$workspaceId/insights" params={params}>
          <IconChartBar aria-hidden />
          {t("shell.nav.insights")}
        </NavLink>
        {isAdmin && (
          <NavLink to="/app/$workspaceId/admin" params={params}>
            <IconShieldCheck aria-hidden />
            {t("shell.nav.admin")}
          </NavLink>
        )}
        <h2 className="m-0 px-2.5 pt-4 pb-1 text-xs font-semibold text-muted">
          {t("settings.title")}
        </h2>
        {settingRows.map(([section, Icon]) => (
          <NavLink key={section} to="." search={openDialog({ settings: section })}>
            <Icon aria-hidden />
            <span className="flex-1">{t(`settings.sections.${section}`)}</span>
            <IconChevronRight aria-hidden />
          </NavLink>
        ))}
        <Button
          className={cn(rowClassName, "mt-3 text-danger [&_svg]:text-danger")}
          onPress={() => {
            logout.mutate({});
          }}
        >
          <IconLogout aria-hidden />
          {t("shell.me.logout")}
        </Button>
      </div>
    </>
  );
};
