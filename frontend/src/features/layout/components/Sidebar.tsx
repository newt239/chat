import { IconSearch } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { NavLink } from "#/components/block/NavLink/NavLink";
import { WorkspaceMenu } from "#/features/workspace/components/WorkspaceMenu";

import { sidebarNavTone } from "../utils/navTone";
import { MiniPlayerSlot } from "./MiniPlayerSlot";
import { NavigationList } from "./NavigationList";
import { SidebarFooter } from "./SidebarFooter";

type SidebarProps = {
  workspaceId: string;
};

export const Sidebar = ({ workspaceId }: SidebarProps) => {
  const { t } = useTranslation();

  return (
    <aside
      aria-label={t("shell.sidebar.label")}
      className={`flex w-[248px] shrink-0 flex-col bg-side text-side-fg ${sidebarNavTone}`}
    >
      <div className="flex h-12 shrink-0 items-center gap-1 pr-2 pl-2.5">
        <WorkspaceMenu workspaceId={workspaceId} />
      </div>
      <NavLink
        to="/app/$workspaceId/search"
        params={{ workspaceId }}
        className="mx-2.5 mb-1.5 h-[30px] w-auto bg-(--nav-hover) text-[13px] text-(--nav-muted)"
      >
        <IconSearch aria-hidden />
        {t("shell.nav.search")}
      </NavLink>
      <NavigationList workspaceId={workspaceId} />
      <MiniPlayerSlot />
      <SidebarFooter />
    </aside>
  );
};
