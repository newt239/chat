import type { ReactNode } from "react";

import { IconChevronRight } from "@tabler/icons-react";
import { MenuItem as AriaMenuItem, Keyboard } from "react-aria-components";

import { cn, overlayStyles } from "#/components/ui/styles/styles";

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
      // href を渡すと a 要素になる
      "no-underline [&_svg]:text-muted data-focused:[&_svg]:text-accent-fg",
      tone === "danger" &&
        "text-danger data-focused:bg-danger data-focused:text-danger-fg [&_svg]:text-danger data-focused:[&_svg]:text-danger-fg",
    )}
  >
    {({ hasSubmenu }) => (
      <>
        {icon}
        <span className="flex-1 truncate">{children}</span>
        {shortcut && <Keyboard className="font-mono text-xs opacity-55">{shortcut}</Keyboard>}
        {hasSubmenu && <IconChevronRight aria-hidden className="size-3.5" />}
      </>
    )}
  </AriaMenuItem>
);
