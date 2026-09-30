import type { ReactNode } from "react";

import { Menu as AriaMenu, Popover, SubmenuTrigger } from "react-aria-components";

import { cn, overlayStyles } from "#/components/ui/styles/styles";

import { MenuItem } from "../MenuItem/MenuItem";

type SubmenuProps = {
  label: string;
  icon: ReactNode;
  // MenuItem / MenuSeparator / MenuSection
  children: ReactNode;
};

// Menu / ContextMenu の中に置く、横に開く入れ子のメニュー
export const Submenu = ({ label, icon, children }: SubmenuProps) => (
  <SubmenuTrigger>
    <MenuItem icon={icon}>{label}</MenuItem>
    <Popover offset={-4} className={cn(overlayStyles.popover, "overflow-y-auto")}>
      <AriaMenu aria-label={label} className={overlayStyles.menu}>
        {children}
      </AriaMenu>
    </Popover>
  </SubmenuTrigger>
);
