import { IconBell, IconHome, IconMessageCircle, IconUser } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { mobileTabAtom } from "#/providers/store/ui";

import { useUnreadSummary } from "../hooks/useUnreadSummary";

type MobileTabBarProps = {
  workspaceId: string;
};

export const MobileTabBar = ({ workspaceId }: MobileTabBarProps) => {
  const { t } = useTranslation();
  const tab = useAtomValue(mobileTabAtom);
  const { activityUnread, dmUnread } = useUnreadSummary(workspaceId);
  const tabs = [
    {
      icon: IconHome,
      label: t("shell.tabs.home"),
      name: "home",
      to: "/app/$workspaceId",
      unread: 0,
    },
    {
      icon: IconMessageCircle,
      label: t("shell.tabs.dms"),
      name: "dms",
      to: "/app/$workspaceId/dms",
      unread: dmUnread,
    },
    {
      icon: IconBell,
      label: t("shell.tabs.activity"),
      name: "activity",
      to: "/app/$workspaceId/activity",
      unread: activityUnread,
    },
    {
      icon: IconUser,
      label: t("shell.tabs.me"),
      name: "me",
      to: "/app/$workspaceId/me",
      unread: 0,
    },
  ] as const;

  return (
    <nav
      aria-label={t("shell.sidebar.label")}
      className="flex shrink-0 border-t border-border bg-surface px-2 pt-2 pb-[max(10px,env(safe-area-inset-bottom))]"
    >
      {tabs.map(({ icon: TabIcon, label, name, to, unread }) => (
        <Link
          key={name}
          to={to}
          params={{ workspaceId }}
          // ホームは他のタブの親のパスなので、完全一致のときだけ現在地にする
          activeOptions={{ exact: name === "home" }}
          data-tab={tab === name ? "on" : "off"}
          className="group flex min-h-14 flex-1 flex-col items-center justify-center gap-1 text-caption text-subtle no-underline data-[tab=on]:font-semibold data-[tab=on]:text-accent-text"
        >
          {/* アクティブなタブはアイコンを薄い色のピルで囲む */}
          <span className="relative grid h-8 w-14 place-items-center rounded-full transition-colors group-data-[tab=on]:bg-accent-soft [&_svg]:size-5.5">
            <TabIcon aria-hidden />
            {unread > 0 && (
              <span className="absolute -top-0.5 left-[calc(50%+6px)] box-content grid h-4 min-w-4 place-items-center rounded-full border-2 border-surface bg-danger px-0.75 text-caption leading-none font-bold text-danger-fg">
                {unread > 99 ? "99+" : unread}
              </span>
            )}
          </span>
          {label}
        </Link>
      ))}
    </nav>
  );
};
