import { IconSearch } from "@tabler/icons-react";
import { useAtom } from "jotai";
import { useTranslation } from "react-i18next";

import { NavLink } from "#/components/block/NavLink/NavLink";
import { sidebarNavTone } from "#/components/block/NavLink/navTone";
import { sidebarWidthRanges, sidebarWidthsAtom } from "#/features/layout/atoms";
import { ResizeHandle } from "#/features/layout/components/ResizeHandle";
import { MiniPlayer } from "#/features/player/components/MiniPlayer";
import { WorkspaceMenu } from "#/features/workspace/components/WorkspaceMenu";

import { NavigationList } from "./NavigationList";
import { SidebarFooter } from "./SidebarFooter";

type SidebarProps = {
  workspaceId: string;
};

export const Sidebar = ({ workspaceId }: SidebarProps) => {
  const { t } = useTranslation();
  const [widths, setWidths] = useAtom(sidebarWidthsAtom);

  return (
    <aside
      aria-label={t("shell.sidebar.label")}
      style={{ width: widths.left }}
      className={`relative flex shrink-0 flex-col bg-side text-side-fg ${sidebarNavTone}`}
    >
      <div className="flex h-12 shrink-0 items-center gap-1 pr-2 pl-2.5">
        <WorkspaceMenu workspaceId={workspaceId} />
      </div>
      <NavLink
        to="/app/$workspaceId/search"
        params={{ workspaceId }}
        className="mx-2.5 mb-1.5 h-7.5 w-auto bg-(--nav-hover) text-body-sm text-(--nav-muted)"
      >
        <IconSearch aria-hidden />
        {t("shell.nav.search")}
      </NavLink>
      <NavigationList workspaceId={workspaceId} />
      <MiniPlayer variant="sidebar" />
      <SidebarFooter />
      <ResizeHandle
        label={t("shell.sidebar.resize")}
        edge="right"
        value={widths.left}
        {...sidebarWidthRanges.left}
        onChange={(left) => {
          setWidths({ ...widths, left });
        }}
      />
    </aside>
  );
};
