import { useAtom } from "jotai";
import { useTranslation } from "react-i18next";

import { surfaceNavTone } from "#/components/ui/styles/navTone";
import { sidebarWidthRanges, sidebarWidthsAtom } from "#/features/layout/atoms";
import { ResizeHandle } from "#/features/layout/components/ResizeHandle";

import { NavigationList } from "./NavigationList";
import { SidebarNavTabs } from "./SidebarNavTabs";

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
      className={`relative flex shrink-0 flex-col rounded-lg bg-surface text-text ${surfaceNavTone}`}
    >
      <SidebarNavTabs workspaceId={workspaceId} />
      <NavigationList workspaceId={workspaceId} />
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
