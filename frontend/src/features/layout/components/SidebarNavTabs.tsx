import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";

import { navShortcuts } from "../utils/navShortcuts";

type SidebarNavTabsProps = {
  workspaceId: string;
};

// チャンネル一覧の上の切り替え。幅が狭いのでアイコンだけを並べ、名前はツールチップと aria-label で伝える
export const SidebarNavTabs = ({ workspaceId }: SidebarNavTabsProps) => {
  const { t } = useTranslation();
  return (
    <nav
      aria-label={t("shell.sidebar.shortcuts")}
      className="mx-2 mt-2 mb-1 grid shrink-0 grid-cols-4 gap-0.5 rounded-lg bg-sunken p-0.5"
    >
      {navShortcuts.map(({ icon: Icon, label, to }) => (
        <Tooltip key={to} content={t(label)}>
          <Link
            to={to}
            params={{ workspaceId }}
            aria-label={t(label)}
            className="flex h-8 items-center justify-center rounded-md text-muted no-underline data-hovered:text-text data-[status=active]:bg-surface data-[status=active]:text-accent-text data-[status=active]:shadow-sm [&_svg]:size-4.5"
          >
            <Icon aria-hidden />
          </Link>
        </Tooltip>
      ))}
    </nav>
  );
};
