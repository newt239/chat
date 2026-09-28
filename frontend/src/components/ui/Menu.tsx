import type { ReactElement, ReactNode } from "react";

import { Menu as AriaMenu, MenuTrigger, Popover } from "react-aria-components";

import { cn, overlayStyles } from "./styles";

import type { PopoverProps } from "react-aria-components";

type MenuProps = {
  // React Aria の Button（IconButton など）。メニューの名前はトリガーのラベルになる
  trigger: ReactElement;
  // MenuItem / MenuSeparator / MenuSection
  children: ReactNode;
  placement?: PopoverProps["placement"];
};

export const Menu = ({ trigger, children, placement = "bottom end" }: MenuProps) => (
  <MenuTrigger>
    {trigger}
    <Popover
      offset={4}
      placement={placement}
      className={cn(overlayStyles.popover, "overflow-y-auto")}
    >
      <AriaMenu className={overlayStyles.menu}>{children}</AriaMenu>
    </Popover>
  </MenuTrigger>
);
