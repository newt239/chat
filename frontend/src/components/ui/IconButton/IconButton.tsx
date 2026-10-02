import type { ReactNode } from "react";

import { Button as AriaButton } from "react-aria-components";

import { focusRing, withBaseClassName } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";

import type { ButtonProps as AriaButtonProps } from "react-aria-components";

type IconButtonProps = Omit<AriaButtonProps, "children" | "aria-label"> & {
  // aria-label とツールチップの両方に使う
  label: string;
  children: ReactNode;
};

// IconToggleButton と共有する見た目
export const iconButtonClassName = `relative inline-grid size-[30px] shrink-0 cursor-pointer place-items-center rounded-md text-muted transition-colors data-disabled:cursor-default data-disabled:text-subtle data-hovered:bg-hover data-hovered:text-text data-pressed:bg-hover [&_svg]:size-[18px] ${focusRing}`;

export const IconButton = ({ label, className, children, ...props }: IconButtonProps) => (
  <Tooltip content={label}>
    <AriaButton
      {...props}
      aria-label={label}
      className={withBaseClassName(className, iconButtonClassName)}
    >
      {children}
    </AriaButton>
  </Tooltip>
);
