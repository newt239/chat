import type { ReactNode } from "react";

import { Button as AriaButton } from "react-aria-components";

import { withBaseClassName } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";

import { iconButtonClassName } from "./iconButtonClassName";

import type { ButtonProps as AriaButtonProps } from "react-aria-components";

type IconButtonProps = Omit<AriaButtonProps, "children" | "aria-label"> & {
  // aria-label とツールチップの両方に使う
  label: string;
  children: ReactNode;
};

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
