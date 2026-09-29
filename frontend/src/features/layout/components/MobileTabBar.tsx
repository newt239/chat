import { IconBell, IconHome, IconMessageCircle, IconUser } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { mobileTabAtom } from "#/providers/store/ui";

import { useUnreadSummary } from "../hooks/useUnreadSummary";

import type { MobileTab } from "#/providers/store/ui";

type MobileTabBarProps = {
  workspaceId: string;
};

const tabClassName =
  "relative flex min-h-[46px] flex-1 flex-col items-center justify-center gap-px text-[10.5px] text-subtle no-underline data-[tab=on]:font-semibold data-[tab=on]:text-accent-text [&_svg]:size-[22px]";

const pip = (count: number) =>
  count > 0 && (
    <span className="absolute top-1 left-[calc(50%+5px)] box-content grid h-4 min-w-4 place-items-center rounded-full border-2 border-surface bg-danger px-[3px] text-[10px] leading-none font-bold text-danger-fg">
      {count > 99 ? "99+" : count}
    </span>
  );

export const MobileTabBar = ({ workspaceId }: MobileTabBarProps) => {
  const { t } = useTranslation();
  const tab = useAtomValue(mobileTabAtom);
  const { activityUnread, dmUnread } = useUnreadSummary(workspaceId);
  const params = { workspaceId };
  const state = (name: MobileTab) => (tab === name ? "on" : "off");

  return (
    <nav
      aria-label={t("shell.sidebar.label")}
      className="flex shrink-0 border-t border-border bg-surface px-2 pt-1 pb-[max(8px,env(safe-area-inset-bottom))]"
    >
      <Link
        to="/app/$workspaceId"
        params={params}
        data-tab={state("home")}
        className={tabClassName}
      >
        <IconHome aria-hidden />
        {t("shell.tabs.home")}
      </Link>
      <Link
        to="/app/$workspaceId/dms"
        params={params}
        data-tab={state("dms")}
        className={tabClassName}
      >
        <IconMessageCircle aria-hidden />
        {pip(dmUnread)}
        {t("shell.tabs.dms")}
      </Link>
      <Link
        to="/app/$workspaceId/activity"
        params={params}
        data-tab={state("activity")}
        className={tabClassName}
      >
        <IconBell aria-hidden />
        {pip(activityUnread)}
        {t("shell.tabs.activity")}
      </Link>
      <Link
        to="/app/$workspaceId/me"
        params={params}
        data-tab={state("me")}
        className={tabClassName}
      >
        <IconUser aria-hidden />
        {t("shell.tabs.me")}
      </Link>
    </nav>
  );
};
