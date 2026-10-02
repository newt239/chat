import type { ReactNode } from "react";

import { ToggleButton } from "react-aria-components";

import { iconButtonClassName } from "#/components/ui/IconButton/iconButtonClassName";
import { cn, withBaseClassName } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";

import type { ToggleButtonProps } from "react-aria-components";

type IconToggleButtonProps = Omit<ToggleButtonProps, "children" | "aria-label"> & {
  // aria-label とツールチップの両方に使う
  label: string;
  children: ReactNode;
};

// 押すたびにオンとオフが切り替わるアイコンのボタン。状態は aria-pressed で伝える
export const IconToggleButton = ({
  label,
  className,
  children,
  ...props
}: IconToggleButtonProps) => (
  <Tooltip content={label}>
    <ToggleButton
      {...props}
      aria-label={label}
      className={withBaseClassName(
        className,
        cn(
          iconButtonClassName,
          "data-selected:bg-accent-soft data-selected:text-accent-text data-selected:data-hovered:bg-accent-soft",
        ),
      )}
    >
      {children}
    </ToggleButton>
  </Tooltip>
);
