import { IconSearch } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { NavLink } from "#/components/block/NavLink/NavLink";
import { mobileNavTone } from "#/components/block/NavLink/navTone";
import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Link } from "#/components/ui/Link/Link";
import { WorkspaceMenu } from "#/features/workspace/components/WorkspaceMenu";
import { useMe } from "#/hooks/useMe";

import { NavigationList } from "./NavigationList";

type MobileHomeProps = {
  workspaceId: string;
};

// モバイルの「ホーム」タブ。サイドバーと同じ一覧を大きめの行で出す
export const MobileHome = ({ workspaceId }: MobileHomeProps) => {
  const { t } = useTranslation();
  const { data: user } = useMe();

  return (
    <div className={`flex min-h-0 flex-1 flex-col text-title font-normal ${mobileNavTone}`}>
      <header className="flex shrink-0 items-center gap-2 px-3 pt-2 pb-1 [&_button]:text-heading [&_button]:font-normal">
        <WorkspaceMenu workspaceId={workspaceId} />
        {user && (
          <Link
            to="/app/$workspaceId/me"
            params={{ workspaceId }}
            aria-label={t("shell.tabs.me")}
            className="no-underline"
          >
            <Avatar name={user.displayName} src={user.avatarUrl} size={30} presence="online" />
          </Link>
        )}
      </header>
      <NavLink
        to="/app/$workspaceId/search"
        params={{ workspaceId }}
        className="mx-3 my-1.5 h-10 w-auto rounded-lg bg-sunken text-muted"
      >
        <IconSearch aria-hidden />
        {t("shell.nav.search")}
      </NavLink>
      <NavigationList workspaceId={workspaceId} />
    </div>
  );
};
