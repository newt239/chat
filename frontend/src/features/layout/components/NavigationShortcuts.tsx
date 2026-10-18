import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { navItemClassName } from "#/components/ui/styles/navTone";

import { navShortcuts } from "../utils/navShortcuts";

type NavigationShortcutsProps = {
  workspaceId: string;
};

export const NavigationShortcuts = ({ workspaceId }: NavigationShortcutsProps) => {
  const { t } = useTranslation();
  return (
    <div className="flex shrink-0 flex-col gap-px px-1.5">
      {navShortcuts.map(({ icon: Icon, label, to }) => (
        <Link key={to} className={navItemClassName} to={to} params={{ workspaceId }}>
          <Icon aria-hidden />
          {t(label)}
        </Link>
      ))}
    </div>
  );
};
