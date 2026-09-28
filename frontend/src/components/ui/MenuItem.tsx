import type { ReactNode } from "react";

import { MenuItem as AriaMenuItem, Keyboard } from "react-aria-components";

import { cn, overlayStyles } from "./styles";

import type { MenuItemProps as AriaMenuItemProps } from "react-aria-components";

type MenuItemProps = Omit<AriaMenuItemProps, "children" | "className"> & {
  children: string;
  icon?: ReactNode;
  shortcut?: string;
  tone?: "default" | "danger";
};

export const MenuItem = ({
  children,
  icon,
  shortcut,
  tone = "default",
  ...props
}: MenuItemProps) => (
  <AriaMenuItem
    textValue={children}
    {...props}
    className={cn(
      overlayStyles.listItem,
      "[&_svg]:text-muted data-focused:[&_svg]:text-accent-fg",
      tone === "danger" &&
        "text-danger data-focused:bg-danger data-focused:text-danger-fg [&_svg]:text-danger data-focused:[&_svg]:text-danger-fg",
    )}
  >
    {icon}
    <span className="flex-1 truncate">{children}</span>
    {shortcut && <Keyboard className="font-mono text-xs opacity-55">{shortcut}</Keyboard>}
  </AriaMenuItem>
);
