import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";

import { useUnreadSummary } from "../hooks/useUnreadSummary";
import { mobileTabs } from "../utils/mobileTabs";

import type { MobileTab } from "../utils/mobileTabs";

type MobileTabBarProps = {
  workspaceId: string;
  tab: MobileTab;
};

export const MobileTabBar = ({ workspaceId, tab }: MobileTabBarProps) => {
  const { t } = useTranslation();
  const { activityUnread, dmUnread } = useUnreadSummary(workspaceId);
  const unreads = { activity: activityUnread, dms: dmUnread, home: 0, me: 0 };

  return (
    <nav
      aria-label={t("shell.sidebar.label")}
      className="flex shrink-0 border-t border-border bg-surface px-2 pt-2 pb-[max(10px,env(safe-area-inset-bottom))]"
    >
      {mobileTabs.map(({ icon: TabIcon, name, to }) => {
        const unread = unreads[name];
        return (
          <Link
            key={name}
            to={to}
            params={{ workspaceId }}
            // ホームは他のタブの親のパスなので、完全一致のときだけ現在地にする
            activeOptions={{ exact: name === "home" }}
            data-tab={tab.name === name ? "on" : "off"}
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
            {t(`shell.tabs.${name}`)}
          </Link>
        );
      })}
    </nav>
  );
};
